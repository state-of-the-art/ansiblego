package modules

import (
	"embed"
	"path"
)

//go:embed task
var modules embed.FS

var modules_cache = map[string]any{}

func InitEmbedded() {
	mtypes, _ := modules.ReadDir(".")
	for _, mtype := range mtypes {
		mods, _ := modules.ReadDir(mtype.Name())
		for _, mod := range mods {
			// If the name contains .go extension - cut it out
			name := mod.Name()
			if path.Ext(name) == ".go" {
				name = name[:len(name)-3]
			}
			modules_cache[path.Join(mtype.Name(), name)] = nil
		}
	}
	// TODO: Enable in debug mode
	//log.Println("Embedded modules:", len(modules_cache))
	//for k, _ := range modules_cache {
	//	log.Println(" ", k)
	//}
}

// Checks if the module is available
func IsTask(name string) bool {
	_, ok := modules_cache["task/"+name]

	return ok
}

func SetCache(typ, name string, val any) {
	p := path.Join(typ, name)
	if modules_cache[p] == nil {
		modules_cache[p] = val
	}
}

func GetCache(typ, name string) (any, bool) {
	p := path.Join(typ, name)
	if modules_cache[p] != nil {
		return modules_cache[p], true
	}
	return nil, false
}
