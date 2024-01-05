package winrm

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/masterzen/winrm"
)

type TransportWinRM struct {
	client *winrm.Client
}

func New(user, pass, host string, port int) (*TransportWinRM, error) {
	tr := &TransportWinRM{}

	if err := tr.connectPassword(user, pass, host, port); err != nil {
		return nil, fmt.Errorf("Failed to connect with password: %v", err)
	}

	if _, _, err := tr.Check(); err != nil {
		return nil, err
	}

	return tr, nil
}

func (tr *TransportWinRM) connectPassword(user, pass, host string, port int) (err error) {
	config := winrm.NewEndpoint(
		host, // Host to connect to
		port, // WinRM port
		true, // Use TLS
		true, // Allow insecure connection
		nil,  // CA certificate
		nil,  // Client Certificate
		nil,  // Client Key
		0,    // Timeout
	)

	tr.client, err = winrm.NewClient(config, user, pass)
	if err != nil {
		return fmt.Errorf("Failed to create WinRM client: %v", err)
	}

	return nil
}

func (tr *TransportWinRM) Execute(cmd string, stdout, stderr io.Writer) (err error) {
	_, err = tr.client.Run(cmd, stdout, stderr)
	if err != nil {
		return fmt.Errorf("Failed to run command: %v", err)
	}
	return nil
}

func (tr *TransportWinRM) ExecuteInput(cmd string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	_, err = tr.client.RunWithInput(cmd, stdout, stderr, stdin)
	if err != nil {
		return fmt.Errorf("Failed to run command: %v", err)
	}
	return nil
}

func (tr *TransportWinRM) Check() (kernel, arch string, err error) {
	// TODO: most of the time it's true, but who knows those
	// poor things who runs winrm server on linux/mac...
	kernel = "windows"

	// Get remote system arch
	arch_buf := bytes.Buffer{}
	stderr_buf := bytes.Buffer{}
	if err = tr.Execute("set processor", &arch_buf, &stderr_buf); err != nil {
		return kernel, arch, fmt.Errorf("Unable to get the remote system arch: %v", err)
	}

	out_lines := strings.Split(arch_buf.String(), "\n")
	for _, l := range out_lines {
		if strings.HasPrefix(strings.TrimSpace(l), "PROCESSOR_ARCHITECTURE=") {
			arch = strings.ToLower(strings.TrimSpace(strings.Split(l, "=")[1]))
		}
	}

	if arch == "" {
		return kernel, arch, fmt.Errorf("No arch found for remote system: %v, %v", arch_buf.String(), stderr_buf.String())
	}

	if arch == "x86_64" || arch == "x64" {
		arch = "amd64"
	}

	return kernel, arch, nil
}

// Rewritten for simplicity version of very quick winrm file copy through stdin
// for io.Reader from: https://github.com/jbrekelmans/go-winrm/blob/master/copier.go
func (tr *TransportWinRM) Copy(content io.Reader, dst string, mode os.FileMode) error {
	// TODO: mode
	script := `begin {
		$path = ` + powerShellQuotedStringLiteral(dst) + `
		$DebugPreference = "Continue"
		$ErrorActionPreference = "Stop"
		Set-StrictMode -Version 2
		$fd = [System.IO.File]::Create($path)
		$sha256 = [System.Security.Cryptography.SHA256CryptoServiceProvider]::Create()
		$bytes = @() #initialize for empty file case
	}
	process {
		$bytes = [System.Convert]::FromBase64String($input)
		$sha256.TransformBlock($bytes, 0, $bytes.Length, $bytes, 0) | Out-Null
		$fd.Write($bytes, 0, $bytes.Length)
	}
	end {
		$sha256.TransformFinalBlock($bytes, 0, 0) | Out-Null
		$hash = [System.BitConverter]::ToString($sha256.Hash).Replace("-", "").ToLowerInvariant()
		$fd.Close()
		Write-Output $hash
	}`

	shell, err := tr.client.CreateShell()
	if err != nil {
		return err
	}
	defer shell.Close()

	cmd, err := shell.Execute(powerShellScript(script))
	if err != nil {
		return err
	}
	defer cmd.Close()

	var wg sync.WaitGroup
	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	stdout_buf := bytes.Buffer{}
	wg.Add(2)
	go copyFunc(&stdout_buf, cmd.Stdout)
	go copyFunc(ioutil.Discard, cmd.Stderr)

	// Preparing sha256 calculate object
	sha256_loc_obj := sha256.New()

	// Making chunk buffer for base64 encoding (which is 4 bytes for 3 bytes of data)
	chunk := make([]byte, (tr.client.Parameters.EnvelopeSize-1000)/4*3)
	b64_chunk := make([]byte, tr.client.Parameters.EnvelopeSize-1000+2)
	id := 0
	for {
		chunk_len, err := content.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if chunk_len == 0 {
			cmd.Stdin.Close()
			break
		}

		sha256_loc_obj.Write(chunk[:chunk_len])

		// Using base64 encoding because it's quite hard to transfer binary data through winrm
		base64.StdEncoding.Encode(b64_chunk, chunk[:chunk_len])
		// 2 additional bytes for CRLF in the end of b64 encoded buffer
		b64_len := base64.StdEncoding.EncodedLen(chunk_len) + 2
		b64_chunk[b64_len-2] = '\r'
		b64_chunk[b64_len-1] = '\n'
		if written_len, err := cmd.Stdin.Write(b64_chunk[:b64_len]); b64_len != written_len || err != nil {
			return fmt.Errorf("Error during copying chunk (bytes: %d, written: %d): %v", b64_len, written_len, err)
		}
		id++
	}

	sha256_loc := hex.EncodeToString(sha256_loc_obj.Sum(nil))

	cmd.Wait()
	wg.Wait()

	if cmd.ExitCode() != 0 {
		return fmt.Errorf("Copy operation returned code=%d", cmd.ExitCode())
	}

	// Verify sha256 to validate the data was copied correctly
	sha256_rem := strings.TrimSpace(stdout_buf.String())
	if sha256_rem != sha256_loc {
		return fmt.Errorf("Copy failed due to checksums mismatch: %v != %v", sha256_rem, sha256_loc)
	}

	return nil
}
