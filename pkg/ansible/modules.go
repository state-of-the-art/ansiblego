package ansible

// This logic allows to cache gomacro interps to not process them twice
// In theory it gives ~30x boost over the regular creating the new interp

import (
	"embed"
	"path"

	"github.com/cosmos72/gomacro/fast"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

//go:embed task
//go:embed fact
var modules embed.FS

var modules_cache = map[string]*fast.Interp{}

func InitEmbeddedModules() {
	mtypes, _ := modules.ReadDir(".")
	for _, mtype := range mtypes {
		mods, _ := modules.ReadDir(mtype.Name())
		for _, mod := range mods {
			// If the name contains .go extension - cut it out
			name := mod.Name()
			if path.Ext(name) == ".go" {
				name = name[:len(name)-3]
			}
			// Init the cache by placing nil there with the name
			modules_cache[path.Join(mtype.Name(), name)] = nil
		}
	}
	if log.Verbosity == log.DEBUG {
		log.Debug("Embedded modules:", len(modules_cache))
		for k, _ := range modules_cache {
			log.Debug(" ", k)
		}
	}
}

// Checks if the module is available
func ModuleIsTask(name string) bool {
	_, ok := modules_cache["task/"+name]

	return ok
}

func ModuleIsFact(name string) bool {
	_, ok := modules_cache["fact/"+name]

	return ok
}

func ModuleSetCache(typ, name string, interp *fast.Interp) {
	p := path.Join(typ, name)
	modules_cache[p] = interp
}

func ModuleGetCache(typ, name string) *fast.Interp {
	p := path.Join(typ, name)
	return modules_cache[p]
}
