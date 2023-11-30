package caps

// Doc: https://github.com/ansible/ansible/blob/stable-2.9/lib/ansible/module_utils/facts/system/caps.py

import (
	"os/exec"
	"strings"
	"time"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
	"github.com/state-of-the-art/ansiblego/pkg/util"
)

func Collect() (data ansible.OrderedMap) {
	caps_enforced := "N/A"
	caps := []string{}

	// Look for capsh in PATH
	capsh_path, err := exec.LookPath("capsh")
	if err != nil {
		log.Debugf("Skipping caps facts collecting: Unable to locate `capsh` in PATH:", err)
		return
	}

	stdout, _, err := util.RunCommand(time.Second, capsh_path, "--print")
	if err != nil {
		log.Debugf("Skipping caps facts collecting: Unable to get `capsh --print` output:", err)
		return
	}

	// Parsing the output trying to find Current state of caps
	for _, line := range strings.Split(stdout, "\n") {
		if len(line) < 1 {
			continue
		}
		if strings.HasPrefix(line, "Current:") {
			if strings.TrimSpace(strings.Split(line, ":")[1]) == "=ep" {
				caps_enforced = "False"
			} else {
				caps_enforced = "True"
				for _, val := range strings.Split(strings.Split(line, "=")[1], ",") {
					caps = append(caps, strings.TrimSpace(val))
				}
			}
			break
		}
	}

	data.Set("system_capabilities", caps)
	data.Set("system_capabilities_enforced", caps_enforced)

	return data
}
