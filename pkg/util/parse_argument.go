package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

// Function is useful to parse command option or argument that contains key=value,
// "-" to read yaml/json stream from stdin, @path to read yaml/json file or json data
// yaml_decoder could be set to nil or used to share context of the read buffer between multiple runs
func ParseArgument(desc, option string, yaml_decoder *yaml.Decoder) (out map[string]any, err error) {
	if option == "-" {
		log.Debugf("Reading %ss yaml/json data from stream...", desc)

		if yaml_decoder == nil {
			yaml_decoder = yaml.NewDecoder(os.Stdin)
		}

		// Read one yaml document out of the stdin stream
		if err := yaml_decoder.Decode(&out); err != nil {
			if err != io.EOF {
				return out, fmt.Errorf("Unable to parse yaml/json from stdin stream: %v", err)
			}
		}
		return out, err
	}
	if strings.HasPrefix(option, "@") {
		log.Debugf("Reading %ss yaml/json file: %q", desc, option[1:])

		data, err := ioutil.ReadFile(option[1:])
		if err != nil {
			return out, fmt.Errorf("Unable to read file %q=%v", option[1:], err)
		}

		if err = yaml.Unmarshal(data, &out); err != nil {
			return out, fmt.Errorf("Unable to parse yaml/json file %q=%v", option[1:], err)
		}
		return out, err
	}
	if strings.HasPrefix(option, "{") && strings.HasSuffix(option, "}") {
		log.Debugf("Parsing %s json data from option...", desc)
		if err = json.Unmarshal([]byte(option), &out); err != nil {
			return out, fmt.Errorf("Unable to parse json option: %v", err)
		}
		return out, err
	}
	kv := strings.SplitN(option, "=", 2)
	if len(kv) < 2 {
		return out, fmt.Errorf("No value provided for %s: %v", desc, kv[0])
	}
	out[kv[0]] = kv[1]
	log.Tracef("Provided extra var: %q=%q", kv[0], kv[1])

	return out, err
}
