package apparmor

// Doc: https://github.com/ansible/ansible/blob/stable-2.9/lib/ansible/module_utils/facts/system/apparmor.py

import (
	"os"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
)

func Collect() (data ansible.OrderedMap) {
	facts := map[string]string{
		"status": "enabled",
	}

	if _, err := os.Stat("/sys/kernel/security/apparmor"); os.IsNotExist(err) {
		facts["status"] = "disabled"
	}

	data.Set("apparmor", facts)

	return data
}
