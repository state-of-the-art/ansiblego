package ansible

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/cosmos72/gomacro/fast"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1Interface interface {
	Run(vars map[string]any) error
	SetData(data OrderedMap) error
	GetData() OrderedMap
}

// Proxy is needed for interface conversion from script to be compiled
// WARNING: the order of func fields in proxy is ABC: https://github.com/cosmos72/gomacro/issues/128
type P_TaskV1Interface struct {
	Object   any
	GetData_ func(_obj_ any) OrderedMap
	Run_     func(_obj_ any, vars map[string]any) error
	SetData_ func(_obj_ any, data OrderedMap) error
}

func (P *P_TaskV1Interface) GetData() OrderedMap {
	return P.GetData_(P.Object)
}
func (P *P_TaskV1Interface) SetData(data OrderedMap) error {
	return P.SetData_(P.Object, data)
}
func (P *P_TaskV1Interface) Run(vars map[string]any) error {
	return P.Run_(P.Object, vars)
}

// TaskDefault is needed to represent a non-implemented task
type TaskV1Default struct {
	Args OrderedMap // Various arguments for the particular task
}

func (t *TaskV1Default) SetData(data OrderedMap) error {
	// TODO
	return nil
}

func (t *TaskV1Default) GetData() OrderedMap {
	// TODO
	var out OrderedMap
	return out
}

func (t *TaskV1Default) Run(vars map[string]any) error {
	// TODO
	return nil
}

func taskInterp(name string) (*fast.Interp, error) {
	// Looking in cache first
	interp := ModuleGetCache("task", name)
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
		"task/%s/type.go",
		"task/%s/module.go",
		"task/%s.go",
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
		log.Debug("Loading task module src:", mod_path)
		_, err = interp.EvalReader(f)
		interp.Comp.Globals.Filepath = "interpreter"
		if err != nil {
			return nil, err
		}

		module_loaded = true
	}
	if !module_loaded {
		return nil, fmt.Errorf("Unable to load module for task '%s'", name)
	}

	// Storing in cache for later usage
	ModuleSetCache("task", name, interp)

	return interp, nil
}

func GetTaskV1(name string) (out TaskV1Interface, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			err = fmt.Errorf("Error during executing the task module '%s': %s", name, pan)
		}
	}()

	interp, err := taskInterp(name)
	if err != nil { // Could be an error after taskInterp panic
		return nil, err
	}

	task_structv, _ := interp.Eval1("sys_modules.TaskV1Interface(&TaskV1{})")
	if err != nil { // Could be an error after interp.Eval1 panic
		return nil, fmt.Errorf("Task '%s' can't convert the struct `TaskV1` to pointer `TaskV1Interface`: %s", name, err)
	}
	if task_structv.Kind() != reflect.Interface {
		return nil, fmt.Errorf("Task '%s' has issues with struct `TaskV1`", name)
	}
	task_struct := task_structv.Interface().(TaskV1Interface)

	return task_struct, nil
}
