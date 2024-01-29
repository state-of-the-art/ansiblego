package copy

// Doc: https://docs.ansible.com/ansible/2.9/modules/copy_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Local path to a file to copy to the remote server.
	Src ansible.TString
	// Influence whether src needs to be transferred or already is present remotely.
	Remote_src ansible.TString
	// Remote absolute path where the file should be copied to.
	Dest ansible.TString `task:",req"`

	// When used instead of src, sets the contents of a file directly to the specified value.
	Content ansible.TString

	// Name of the user that should own the file/directory, as would be fed to chown.
	Owner ansible.TString
	// Name of the group that should own the file/directory, as would be fed to chown.
	Group ansible.TString
	// The permissions of the destination file or directory.
	Mode ansible.TString
	// When doing a recursive copy set the mode for the directories.
	Directory_mode ansible.TString

	// The attributes the resulting file or directory should have.
	//Attributes    string `,alias:attr`
	// Create a backup file including the timestamp information so you can get the original file back if you somehow clobbered it incorrectly.
	//Backup        bool
	// This option controls the autodecryption of source files using vault.
	//Decrypt       bool   `,def:true`
	// Influence whether the remote file must always be replaced.
	//Force         bool   `,def:true,alias:thirsty`
	// This flag indicates that filesystem links in the destination, if they exist, should be followed.
	//Follow        bool
	// This flag indicates that filesystem links in the source tree, if they exist, should be followed.
	//Local_follow  bool   `,def:true`
	// Influence when to use atomic operation to prevent data corruption or inconsistent reads from the target file.
	//Unsafe_writes bool
	// The validation command to run before copying into place.
	//Validate      bool

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
	d, ok := data.Pop("copy")
	if !ok {
		return fmt.Errorf("Unable to find the 'copy' map in task data")
	}
	fmap, ok := d.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'copy' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("copy", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement copy.Run")

	return
}
