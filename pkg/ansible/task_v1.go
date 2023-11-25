package ansible

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime/debug"

	"github.com/cosmos72/gomacro/fast"
	"github.com/cosmos72/gomacro/imports"

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

func taskInterp() *fast.Interp {
	// TODO: Much more efficient will be to store interpreter as a variable
	// somewhere in modules, but let's leave it for the later optimization.
	interp := fast.New()

	// Discard interp warnings
	if log.Verbosity < log.DEBUG {
		interp.Comp.Globals.Output.Stderr = ioutil.Discard
	}

	// Import logging
	imports.Packages["github.com/state-of-the-art/ansiblego/pkg/log"] = imports.Package{
		Binds: map[string]reflect.Value{
			"Trace":  reflect.ValueOf(log.Trace),
			"Tracef": reflect.ValueOf(log.Tracef),
			"Debug":  reflect.ValueOf(log.Debug),
			"Debugf": reflect.ValueOf(log.Debugf),
			"Info":   reflect.ValueOf(log.Info),
			"Infof":  reflect.ValueOf(log.Infof),
			"Warn":   reflect.ValueOf(log.Warn),
			"Warnf":  reflect.ValueOf(log.Warnf),
			"Error":  reflect.ValueOf(log.Error),
			"Errorf": reflect.ValueOf(log.Errorf),
		},
		Types:    map[string]reflect.Type{},
		Proxies:  map[string]reflect.Type{},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}
	interp.ImportPackage("sys_modules", "github.com/state-of-the-art/ansiblego/pkg/log")

	// Import the TaskV1 interface
	imports.Packages["github.com/state-of-the-art/ansiblego/pkg/ansible"] = imports.Package{
		Binds: map[string]reflect.Value{
			"TaskV1SetData": reflect.ValueOf(TaskV1SetData),
			"TaskV1GetData": reflect.ValueOf(TaskV1GetData),
			"ToYaml":        reflect.ValueOf(ToYaml),
		},
		Types: map[string]reflect.Type{
			"Task":            reflect.TypeOf((*Task)(nil)).Elem(),
			"TaskV1Interface": reflect.TypeOf((*TaskV1Interface)(nil)).Elem(),
			"OrderedMap":      reflect.TypeOf((*OrderedMap)(nil)).Elem(),
		},
		Proxies: map[string]reflect.Type{
			"TaskV1Interface": reflect.TypeOf((*P_TaskV1Interface)(nil)).Elem(),
		},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}
	// The module could not use the import, but we still need it for proper interfacing
	interp.ImportPackage("sys_modules", "github.com/state-of-the-art/ansiblego/pkg/ansible")

	return interp
}

func evalTask(interp *fast.Interp, name string) error {
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
		log.Debug("Loading module src:", mod_path)
		_, err = interp.EvalReader(f)
		interp.Comp.Globals.Filepath = "interpreter"
		if err != nil {
			return err
		}

		module_loaded = true
	}
	if !module_loaded {
		return fmt.Errorf("Unable to load module for task '%s'", name)
	}
	return nil
}

func GetTaskV1(name string) (out TaskV1Interface, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			err = fmt.Errorf("Error during executing the task module '%s': %s\n%s", name, pan, string(debug.Stack()))
		}
	}()

	if val, ok := ModuleGetCache("task", name); ok {
		if out, ok = val.(TaskV1Interface); ok {
			return out, nil
		} else {
			log.Warn("Incorrect task type in cache:", reflect.TypeOf(val))
		}
	}

	interp := taskInterp()

	err = evalTask(interp, name)
	if err != nil { // Could be an error after evalTask panic
		return nil, err
	}

	task_structv, xtype := interp.Eval1("sys_modules.TaskV1Interface(&TaskV1{})")
	if err != nil { // Could be an error after interp.Eval1 panic
		return nil, fmt.Errorf("Task '%s' can't convert the struct `TaskV1` to pointer `TaskV1Interface`: %s", name, err)
	}
	if task_structv.Kind() != reflect.Interface {
		return nil, fmt.Errorf("Task '%s' has issues with struct `TaskV1`", name)
	}
	ModuleSetCache("task", name, xtype)
	task_struct := task_structv.Interface().(TaskV1Interface)

	return task_struct, nil
}
