package inventory

import (
	"os"

	"github.com/relex/aini"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type Inventory = aini.InventoryData
type Host = aini.Host

// Input is a file with inventory data or a list of hosts
func New(input []string) (*Inventory, error) {
	out := &Inventory{}
	for _, host := range input {
		if _, err := os.Stat(host); err == nil {
			// File path that contains inventory - loading it and return
			out, err = ReadIniFile(host)
			if err != nil {
				return nil, log.Errorf("Unable to load inventory file by ini parser: %s", host)
			}
			log.Debugf("Ini data: %q", out)
			return out, nil
		} else {
			// List of hosts
			log.Error("TODO: list of hosts not supported, to be implemented")
		}
	}

	return out, nil
}

func ReadIniFile(path string) (inv *Inventory, err error) {
	// Open and parse
	return aini.ParseFile(path)
}
