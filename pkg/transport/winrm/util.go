package winrm

import (
	"encoding/base64"
	"strings"

	"golang.org/x/text/encoding/unicode"
)

// FormatPowerShellScriptCommandLine returns the command and arguments to run the specified PowerShell script.
// The returned slice contains the following elements:
// PowerShell -NoProfile -NonInteractive -ExecutionPolicy Unrestricted -EncodedCommand <base64>
func powerShellScript(script string) string {
	// Encode string to UTF16-LE
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	encoded, err := encoder.String(script)
	if err != nil {
		return ""
	}

	// Finally make it base64 encoded which is required for powershell.
	script_encoded := base64.StdEncoding.EncodeToString([]byte(encoded))
	return "PowerShell -NoProfile -NonInteractive -ExecutionPolicy Unrestricted -EncodedCommand " + script_encoded
}

func powerShellQuotedStringLiteral(v string) string {
	var sb strings.Builder
	_, _ = sb.WriteRune('\'')
	for _, rune := range v {
		var esc string
		switch rune {
		case '\n':
			esc = "`n"
		case '\r':
			esc = "`r"
		case '\t':
			esc = "`t"
		case '\a':
			esc = "`a"
		case '\b':
			esc = "`b"
		case '\f':
			esc = "`f"
		case '\v':
			esc = "`v"
		case '"':
			esc = "`\""
		case '\'':
			esc = "`'"
		case '`':
			esc = "``"
		case '\x00':
			esc = "`0"
		default:
			_, _ = sb.WriteRune(rune)
			continue
		}
		_, _ = sb.WriteString(esc)
	}
	_, _ = sb.WriteRune('\'')
	return sb.String()
}
