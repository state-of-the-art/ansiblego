package ansible

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/cosmos72/gomacro/fast"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

func factInterp(name string) (*fast.Interp, error) {
	// Looking in cache first
	interp := ModuleGetCache("fact", name)
	if interp != nil {
		return interp, nil
	}

	// Creating new interp
	interp = fast.New()

	// Discard interp warnings
	if log.Verbosity < log.DEBUG {
		interp.Comp.Globals.Output.Stderr = ioutil.Discard
	}

	// Allow just the known gomacro imports.Packages to be imported
	// We need to make sure modules will not fail just because internet is not
	// available or there is a minimal environment without a way to run `go get`.
	// TODO: https://github.com/cosmos72/gomacro/pull/152
	//interp.Comp.CompGlobals.Importer.BlockExternal = true

	// The module could not use the import, but we still need it for proper interfacing
	interp.ImportPackage("sys_modules", "github.com/state-of-the-art/ansiblego/pkg/ansible")

	paths := []string{
		"fact/%s/type.go",
		"fact/%s/module.go",
		"fact/%s.go",
	}
	module_loaded := false
	for _, path := range paths {
		// Check if it has entry file module
		mod_path := fmt.Sprintf(path, name)
		f, err := modules.Open(mod_path)
		if err != nil {
			continue
		}

		interp.Comp.Globals.Filepath = mod_path
		log.Debug("Loading fact module src:", mod_path)
		_, err = interp.EvalReader(f)
		interp.Comp.Globals.Filepath = "interpreter"
		if err != nil {
			return nil, err
		}

		module_loaded = true
	}
	if !module_loaded {
		return nil, fmt.Errorf("Unable to load module for fact '%s'", name)
	}

	// Storing in cache for later usage
	ModuleSetCache("fact", name, interp)

	return interp, nil
}

func CollectV1(name string) (out *OrderedMap, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			err = fmt.Errorf("Error during executing the fact module '%s': %s", name, pan)
		}
	}()

	interp, err := factInterp(name)
	if err != nil { // Could be an error after factInterp panic
		return nil, err
	}

	fact_data, _ := interp.Eval1("Collect()")
	if err != nil { // Could be an error after interp.Eval1 panic
		return nil, fmt.Errorf("Fact '%s' can't execute Collect(): %s", name, err)
	}
	if fact_data.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Fact '%s' has issues with return from `Collect()` kind: %q", name, fact_data.Kind())
	}
	fact_map := fact_data.Interface().(OrderedMap)

	return &fact_map, nil
}
