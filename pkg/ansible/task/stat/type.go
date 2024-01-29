package stat

// Doc: https://docs.ansible.com/ansible/2.9/modules/stat_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Path to the file being managed.
	Path ansible.TString `task:",req,alias:dest,alias:name"`

	// Algorithm to determine checksum of file.
	Checksum_algorithm ansible.TString `task:",alias:checksum,alias:checksum_algo,def:sha1,list:md5 sha1 sha224 sha256 sha384 sha512"`

	// Whether to follow symlinks.
	//Follow bool
	// Get file attributes using lsattr tool if present.
	//Get_attributes bool `task:",def:true,alias:attr,alias:attributes"`
	// Whether to return a checksum of the file.
	//Get_checksum bool `task:",def:true"`
	// Use file magic and return data about the nature of the file.
	//Get_mime bool `task:",def:true,alias:mime,alias:mime_type,alias:mime-type"`
}

// Here the fields comes as complete values never as jinja2 templates
func (t *TaskV1) SetData(data *ansible.OrderedMap) error {
	d, ok := data.Pop("stat")
	if !ok {
		return fmt.Errorf("Unable to find the 'stat' map in task data")
	}
	fmap, ok := d.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'stat' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("stat", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement stat.Run")

	return
}
