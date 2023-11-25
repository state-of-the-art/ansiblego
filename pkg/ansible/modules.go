package ansible

import (
	"embed"
	"path"

	"github.com/cosmos72/gomacro/xreflect"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

//go:embed task
var modules embed.FS

var modules_cache = map[string]*xreflect.Type{}

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

func ModuleSetCache(typ, name string, xtype xreflect.Type) {
	p := path.Join(typ, name)
	if modules_cache[p] == nil {
		modules_cache[p] = &xtype
	}
}

func ModuleGetCache(typ, name string) (any, bool) {
	p := path.Join(typ, name)
	if modules_cache[p] != nil {
		val := xreflect.New(*modules_cache[p])
		return val.Interface(), true
	}
	return nil, false
}
