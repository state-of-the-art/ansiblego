package apt

// Doc: https://docs.ansible.com/ansible/2.9/modules/apt_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// A list of package names, like foo, or package specifier with version, like foo=1.0. Name wildcards (fnmatch) like apt* and version wildcards like foo=1.0* are also supported.
	Name []string `task:",alias:package,alias:pkg"`

	// Run the equivalent of apt-get update before the operation.
	Update_cache bool

	// Indicates the desired package state.
	//State string `task:",def:present,list:absent build-dep latest present fixed"`

	// Ignore if packages cannot be authenticated.
	//Allow_unauthenticated bool
	// Update the apt cache if its older than the cache_valid_time. This option is set in seconds.
	//Cache_valid_time int
	// Path to a .deb package on the remote machine.
	//Deb string
	// Corresponds to the -t option for apt and sets pin priorities
	//Default_release string
	// Add dpkg options to apt command. Defaults to '-o "Dpkg::Options::=--force-confdef" -o "Dpkg::Options::=--force-confold"'
	//Dpkg_options string `task:",def:force-confdef,force-confold"`
	// Corresponds to the --force-yes to apt-get and implies allow_unauthenticated: yes
	//Force bool
	// Force usage of apt-get instead of aptitude
	//Force_apt_get bool
	// Corresponds to the --no-install-recommends option for apt.
	//Install_recommends *bool `task:",alias:install-recommends"`
	// Only upgrade a package if it is already installed.
	//Only_upgrade bool
	// Force the exit code of /usr/sbin/policy-rc.d.
	//Policy_rc_d *int

	// If yes or safe, performs an aptitude safe-upgrade. If full, performs an aptitude full-upgrade. If dist, performs an apt-get dist-upgrade.
	//Upgrade string `task:",def:no,list:dist full no safe yes"`

	// If yes, cleans the local repository of retrieved package files that can no longer be downloaded.
	//Autoclean bool
	// If yes, remove unused dependency packages for all module states except build-dep.
	//Autoremove bool
	// Will force purging of configuration files if the module state is set to absent.
	//Purge bool
}

func (t *TaskV1) SetData(data ansible.OrderedMap) error {
	apt_data, ok := data.Get("apt")
	if !ok {
		return fmt.Errorf("Unable to find the 'apt' map in task data")
	}
	fmap, ok := apt_data.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'apt' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("apt", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) error {
	log.Error("TODO: Implement apt.Run")

	return nil
}
