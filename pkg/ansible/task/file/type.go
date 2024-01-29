package file

// Doc: https://docs.ansible.com/ansible/2.9/modules/file_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Path of the file to link to.
	Src ansible.TString
	// Path to the file being managed.
	Path ansible.TString `task:",req,alias:dest,alias:name"`
	// State of the file in the end
	State ansible.TString `task:",def:file,list:absent directory file hard link touch"`

	// Name of the user that should own the file/directory, as would be fed to chown.
	Owner ansible.TString
	// Name of the group that should own the file/directory, as would be fed to chown.
	Group ansible.TString
	// The permissions the resulting file or directory should have.
	Mode ansible.TString

	// Recursively set the specified file attributes on directory contents.
	Recurse ansible.TBool
	// This flag indicates that filesystem links, if they exist, should be followed.
	Follow ansible.TBool `task:",def:true"`

	// This parameter indicates the time the file's access time should be set to.
	//Access_time        string
	// When used with access_time, indicates the time format that must be used.
	//Access_time_format string `,def:%Y%m%d%H%M.%S`

	// This parameter indicates the time the file's modification time should be set to.
	//Modification_time        string
	// When used with modification_time, indicates the time format that must be used.
	//Modification_time_format string `,def:%Y%m%d%H%M.%S`

	// The attributes the resulting file or directory should have.
	//Attributes    string `,alias:attr`
	// Force the creation of the symlinks in two cases: the source file does not exist (but will appear later); the destination exists and is a file (so, we need to unlink the path file and create symlink to the src file in place of it).
	//Force         bool   `,def:false`
	// Influence when to use atomic operation to prevent data corruption or inconsistent reads from the target file.
	//Unsafe_writes bool

	// The level part of the SELinux file context.
	//Selevel string `,def:s0`
	// The role part of the SELinux file context.
	//Serole  string
	// The type part of the SELinux file context.
	//Setype  string
	// The user part of the SELinux file context.
	//Seuser  string
}

// Here the fields comes as complete values never as jinja2 templates
func (t *TaskV1) SetData(data *ansible.OrderedMap) error {
	d, ok := data.Pop("file")
	if !ok {
		return fmt.Errorf("Unable to find the 'file' map in task data")
	}
	fmap, ok := d.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'file' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("file", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement file.Run")

	return
}
