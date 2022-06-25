package winrm

// This winrm copy utility is a rewrite of the logic from:
// https://github.com/packer-community/winrmcp/blob/master/winrmcp/cp.go
//
// Uses the usual parallel chunk-based approach, but could be optimized
// TODO: through PSRP streaming or https://github.com/jbrekelmans/go-winrm

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/masterzen/winrm"
)

// If you increase this number - change the `uploadChunkWorker` to
// fit the amount of digits to properly order the chunks in directory
const CHUNKS_BATCH = 500

type WinRMCP struct {
	client      *winrm.Client
	num_workers int

	jobs   chan *Job
	errors chan *Job

	workers_done sync.WaitGroup

	combine_sync sync.Mutex
}

type Job struct {
	Id    int
	Batch int
	Data  string
}

func NewWinRMCP(client *winrm.Client, num_workers int) *WinRMCP {
	if num_workers <= 0 {
		num_workers = 1
	}
	return &WinRMCP{
		client:      client,
		num_workers: num_workers,
	}
}

func (wcp *WinRMCP) Copy(content io.Reader, to_path string) error {
	wcp.errors = make(chan *Job, wcp.num_workers)
	defer close(wcp.errors)

	temp_file_name := fmt.Sprintf("winrmcp-%08x", rand.Uint32())
	temp_path := "$env:TEMP\\" + temp_file_name

	// Set the number of parallel winrm operations to maximum
	command := `winrm set winrm/config/service @{MaxConcurrentOperationsPerUser="4294967295"}`
	if _, err := wcp.client.Run(command, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("Unable to set the maximum parallel operations: %v", err)
	}

	log.Printf("Copying file to %s\n", temp_path)

	if err := wcp.uploadChunks("%TEMP%\\"+temp_file_name, content); err != nil {
		return fmt.Errorf("Error uploading file to %s: %v", temp_path, err)
	}

	log.Printf("Moving file from %s to %s", temp_path+".tmp", to_path)

	if err := wcp.restoreContent(temp_path+".tmp", to_path); err != nil {
		return fmt.Errorf("Error restoring file from %s to %s: %v", temp_path, to_path, err)
	}

	log.Printf("Removing temporary file %s", temp_path)

	if err := wcp.cleanupContent(temp_path + ".tmp"); err != nil {
		return fmt.Errorf("Error removing temporary file %s: %v", temp_path, err)
	}

	return nil
}

func (wcp *WinRMCP) uploadChunkWorker(wid int, file_path string) {
	defer wcp.workers_done.Done()

	shell, err := wcp.client.CreateShell()
	if err != nil {
		wcp.errors <- &Job{Id: -1, Batch: -1, Data: fmt.Sprintf("Worker #%d couldn't create shell: %v", wid, err)}
		return
	}
	defer shell.Close()

	for j := range wcp.jobs {
		// Here in the chunk file name we use the general digit ordering
		// in the folder to properly combine the chunks later so it should
		// be the same as number of digits of CHUNKS_BATCH constant
		command := fmt.Sprintf(`echo %s> "%s-%d/%03d.tmp"`, j.Data, file_path, j.Batch, j.Id)
		if err = execCommand(shell, command); err != nil {
			j.Data = fmt.Sprintf("Worker #%d couldn't upload the chunk: %v", wid, err)
			wcp.errors <- j
		}
	}
}

func (wcp *WinRMCP) uploadChunks(file_path string, reader io.Reader) error {
	// We using queue size of the amount of workers
	wcp.jobs = make(chan *Job, wcp.num_workers)

	for w := 1; w <= wcp.num_workers; w++ {
		go wcp.uploadChunkWorker(w, file_path)
		wcp.workers_done.Add(1)
	}

	// Upload the file in chunks to get around the Windows command line size limit.
	// Base64 encodes each set of three bytes into four bytes. In addition the output
	// is padded to always be a multiple of four.
	//
	//   ceil(n / 3) * 4 = m1 - m2
	//
	//   where:
	//     n  = bytes
	//     m1 = max (8192 character command limit.)
	//     m2 = len(file_path)

	chunk_size := ((8000 - len(file_path)) / 4) * 3
	chunk := make([]byte, chunk_size)

	// Read the content and feed the workers
	id := 0
	batch := 0
	for {
		n, err := reader.Read(chunk)

		if err != nil && err != io.EOF {
			close(wcp.jobs)
			return err
		}
		if n == 0 {
			break
		}

		// Non-blocking check for errors in goroutines
		select {
		case job := <-wcp.errors:
			close(wcp.jobs)
			// TODO: get all the errors from the channel
			return fmt.Errorf("Errors occured during chunks copying: %s", job.Data)
		default:
		}

		// Switch batch if an amount of chunks exceed the limit
		if id >= CHUNKS_BATCH {
			batch++
			id = 0
		}

		if id == 0 {
			// Creating new chunk batch directory
			chunk_dir := fmt.Sprintf("%s-%d", file_path, batch)
			if err := wcp.createDirectory(chunk_dir); err != nil {
				close(wcp.jobs)
				return fmt.Errorf("Unable to create tmp chunk batch directory: %v", err)
			}
		}

		// Creating job and placing it in queue
		content := base64.StdEncoding.EncodeToString(chunk[:n])
		wcp.jobs <- &Job{Id: id, Batch: batch, Data: string(content)}

		// Combine a previous batch of completed chunks to the combined file for further processing
		// The interesting part here is that we're waiting for the current branch to fill the queue
		// by a num_workers - which means the previous batch was actually completed
		if batch > 0 && id == wcp.num_workers {
			go wcp.combineChunks(file_path, batch-1)
		}

		id++
	}
	close(wcp.jobs)

	// Wait for all the jobs are completed
	fmt.Println("DEBUG: Wait for jobs to complete")
	wcp.workers_done.Wait()

	// When all the jobs are done we combining the last part of chunks
	if id <= wcp.num_workers {
		// Processing the previous batch
		if err := wcp.combineChunks(file_path, batch-1); err != nil {
			return err
		}
	}
	if id >= 0 {
		// Processing last chunks
		if err := wcp.combineChunks(file_path, batch); err != nil {
			return err
		}
	}

	return nil
}

func (wcp *WinRMCP) combineChunks(file_path string, batch int) error {
	// Make sure combining of chunk batches will go in order
	wcp.combine_sync.Lock()
	defer wcp.combine_sync.Unlock()

	fmt.Printf("Combining of chunks from batch %d\n", batch)

	shell, err := wcp.client.CreateShell()
	if err != nil {
		wcp.errors <- &Job{Id: -1, Batch: batch, Data: fmt.Sprintf("Can't combine chunks: %v", err)}
		return err
	}
	defer shell.Close()

	commands := []string{
		// Make sure the output file exists
		fmt.Sprintf(`type nul >> "%s.tmp"`, file_path),
		// Concatting the output file with the chunk files into a new output file
		fmt.Sprintf(`copy "%s.tmp"+"%s-%d"\*.tmp "%s.tmp2"`, file_path, file_path, batch, file_path),
		// Moving the new output file to the old one place
		fmt.Sprintf(`move /Y "%s.tmp2" "%s.tmp"`, file_path, file_path),
		// Removing the chunks batch directory
		fmt.Sprintf(`rmdir /Q /S "%s-%d"`, file_path, batch),
	}
	if err = execCommand(shell, strings.Join(commands, " && ")); err != nil {
		wcp.errors <- &Job{Id: -1, Batch: batch, Data: fmt.Sprintf("Can't combine chunks: %v", err)}
		return err
	}

	return nil
}

func (wcp *WinRMCP) createDirectory(dir_path string) error {
	fmt.Println("Creating directory:", dir_path)
	shell, err := wcp.client.CreateShell()
	if err != nil {
		return err
	}
	defer shell.Close()

	command := fmt.Sprintf(`mkdir "%s"`, dir_path)
	if err = execCommand(shell, command); err != nil {
		return err
	}

	return nil
}

func (wcp *WinRMCP) restoreContent(fromPath, toPath string) error {
	shell, err := wcp.client.CreateShell()
	if err != nil {
		return err
	}
	defer shell.Close()

	script := fmt.Sprintf(`
		$tmp_file_path = [System.IO.Path]::GetFullPath("%s")
		$dest_file_path = [System.IO.Path]::GetFullPath("%s".Trim("'"))
		if (Test-Path $dest_file_path) {
			if (Test-Path -Path $dest_file_path -PathType container) {
				Exit 1
			} else {
				rm $dest_file_path
			}
		}
		else {
			$dest_dir = ([System.IO.Path]::GetDirectoryName($dest_file_path))
			New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path $dest_dir | Out-Null
		}
		if (Test-Path $tmp_file_path) {
			$reader = [System.IO.File]::OpenText($tmp_file_path)
			$writer = [System.IO.File]::OpenWrite($dest_file_path)
			try {
				for(;;) {
					$base64_line = $reader.ReadLine()
					if ($base64_line -eq $null) { break }
					$bytes = [System.Convert]::FromBase64String($base64_line)
					$writer.write($bytes, 0, $bytes.Length)
				}
			}
			finally {
				$reader.Close()
				$writer.Close()
			}
		} else {
			echo $null > $dest_file_path
		}
	`, fromPath, toPath)

	cmd, err := shell.Execute(winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return fmt.Errorf("restore operation returned code=%d", cmd.ExitCode())
	}
	return nil
}

func (wcp *WinRMCP) cleanupContent(file_path string) error {
	shell, err := wcp.client.CreateShell()
	if err != nil {
		return err
	}
	defer shell.Close()

	script := fmt.Sprintf(`
		$tmp_file_path = [System.IO.Path]::GetFullPath("%s")
		if (Test-Path $tmp_file_path) {
			Remove-Item $tmp_file_path -ErrorAction SilentlyContinue
		}
	`, file_path)

	cmd, err := shell.Execute(winrm.Powershell(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return fmt.Errorf("cleanup operation returned code=%d", cmd.ExitCode())
	}
	return nil
}

func execCommand(shell *winrm.Shell, command string) error {
	cmd, err := shell.Execute(command)
	if err != nil {
		return err
	}

	defer cmd.Close()
	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	wg.Add(2)
	go copyFunc(os.Stdout, cmd.Stdout)
	go copyFunc(os.Stderr, cmd.Stderr)

	cmd.Wait()
	wg.Wait()

	fmt.Println("Command executed:", command[:20], cmd.ExitCode())
	if cmd.ExitCode() != 0 {
		return fmt.Errorf("Execute command returned exit code=%d", cmd.ExitCode())
	}

	return nil
}
