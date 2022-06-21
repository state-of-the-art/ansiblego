package embedbin

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/ulikunitz/xz"
)

// Embedded binary TOKEN format:
// \n--- EMBEDDED_BINARY <kernel>-<arch> <package> ---\n
// examples:
//   kernel:
//     - linux
//     - darwin
//     - windows
//   arch:
//     - amd64
//     - arm64
//   package:
//     - raw (linux-amd64: 31M)
//     - upx (linux-amd64: 8.4M)
//     - gz (linux-amd64: 11M)
//     - xz (linux-amd64: 7.8M)

const (
	// Token is split in 2 parts to not find it accidentally
	TOKEN_PT1         = "\n--- "
	TOKEN_PT2         = "EMBEDDED_BINARY "
	TOKEN_END         = " ---\n"
	HEADER_MAX_LENGTH = 128 // Maximum length of the header in bytes
)

// Will return the found embedded binary or empty array
// TODO: not a great implementation since need to store data
// in memory, but could be optimized later when will be used
func GetEmbeddedBinary(kernel, arch string) ([]byte, error) {
	exec_path, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("Unable to locate the executable path: %v", err)
	}

	// Open the executable for read
	exec_file, err := os.Open(exec_path)
	if err != nil {
		return nil, fmt.Errorf("Unable to locate the executable path: %v", err)
	}
	defer exec_file.Close()

	buf := make([]byte, 8192)
	token := []byte(TOKEN_PT1 + TOKEN_PT2)

	// Special case if the needed kernel/arch is the same as runtime
	if runtime.GOOS == kernel && runtime.GOARCH == arch {
		// Scanning for token and copying the buffer to output
		var out []byte
		for {
			length, err := exec_file.Read(buf)
			if err == io.EOF {
				// We reached the end of the file so stop here
				break
			} else if err != nil {
				return nil, fmt.Errorf("Unable to read bytes: %v", err)
			}
			out = append(out, buf[:length]...)

			// Checking the last added data for token keeping in mind a possible overlap
			sniff_start := len(out) - len(buf) - len(token)
			if sniff_start < 0 {
				sniff_start = 0
			}
			token_pos := bytes.Index(out[sniff_start:], token)
			if token_pos < 0 {
				// No token found so continue
				continue
			}

			// We found the token so can cut the out and break the loop
			out = out[:sniff_start+token_pos]
			break
		}
		return out, nil
	}

	// Locating the required section from the bottom of the executable
	// with overlapping buffers to not miss the token in binary file.
	// Reading it from the bottom is discussable, because basically we
	// need to reread the parsed section again, but we skipping ~30MB
	// of current executable and that could worth it.
	var header_fields []string // The header content separated by space
	var binary_start int64     // Where to start reading of the embedded binary
	binary_end, err := exec_file.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, fmt.Errorf("Unable to get file size: %v", err)
	}

	read_jump := int64(len(buf) - len(token))
	for i := binary_end - int64(len(buf)); i > 0; i -= read_jump {
		file_pos, err := exec_file.Seek(i, os.SEEK_SET)
		if err != nil {
			return nil, fmt.Errorf("Unable to seek position %d: %v", i, err)
		}

		if n, err := exec_file.Read(buf); err != nil {
			return nil, fmt.Errorf("Unable to read bytes on position %d (%d): %d %v", i, file_pos, n, err)
		}

		token_pos := bytes.LastIndex(buf, token)
		if token_pos < 0 {
			// No token found so continue
			continue
		}

		// Processing the token and header
		file_pos, err = exec_file.Seek(file_pos+int64(token_pos), os.SEEK_SET)
		if err != nil {
			return nil, fmt.Errorf("Unable to re-seek position %d: %v", i, err)
		}
		if n, err := exec_file.Read(buf); err != nil {
			return nil, fmt.Errorf("Unable to re-read bytes on position %d: %d %v", token_pos, n, err)
		}
		// The header max size is HEADER_MAX_LENGTH bytes
		token_end_pos := bytes.Index(buf[:HEADER_MAX_LENGTH], []byte(TOKEN_END))
		if token_end_pos < 0 {
			// It was a false alarm
			continue
		}
		token_end_pos += len(TOKEN_END)
		header_fields = strings.Split(string(buf[len(token):token_end_pos-len(TOKEN_END)]), " ")

		// Procesing the found header fields
		if len(header_fields) >= 2 && header_fields[0] == (kernel+"-"+arch) {
			// The required binary was found
			binary_start = file_pos + int64(token_end_pos)
			break
		} else {
			// It's not the binary we need, so continue the search
			// Using i here because the token_pos was found in the first buffer
			binary_end = i + int64(token_pos)
		}
	}
	if binary_start <= 0 || binary_start >= binary_end {
		return nil, fmt.Errorf("Unable to find the required embedded binary in '%s': %s-%s", exec_path, kernel, arch)
	}

	// Set where the binary starts and limit the area of the reader to not read too much
	if _, err := exec_file.Seek(binary_start, os.SEEK_SET); err != nil {
		return nil, fmt.Errorf("Unable to seek position %d: %v", binary_start, err)
	}
	reader := io.LimitReader(exec_file, binary_end-binary_start)

	// If the binary can be executed directly - just return the data
	// or if it's packed - need to unpack it
	if header_fields[1] == "raw" || header_fields[1] == "upx" {
		return io.ReadAll(reader)
	}

	switch header_fields[1] {
	case "raw", "upx":
		return io.ReadAll(reader)
	case "gz":
		r, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to create Gzip reader: %v", err)
		}
		return io.ReadAll(r)
	case "xz":
		r, err := xz.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to create XZ reader: %v", err)
		}
		return io.ReadAll(r)
	}

	return nil, fmt.Errorf("Unsupported packer for embedded binary: %s", header_fields[1])
}
