package ansible

import (
	"reflect"

	"github.com/cosmos72/gomacro/imports"

	"github.com/state-of-the-art/ansiblego/pkg/log"
	"github.com/state-of-the-art/ansiblego/pkg/template"
	"github.com/state-of-the-art/ansiblego/pkg/util"
)

// Loading the usable packages for gomacro modules
func init() {
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

	// Import util
	imports.Packages["github.com/state-of-the-art/ansiblego/pkg/util"] = imports.Package{
		Binds: map[string]reflect.Value{
			"RunCommand":      reflect.ValueOf(util.RunCommand),
			"RunCommandRetry": reflect.ValueOf(util.RunCommandRetry),
		},
		Types:    map[string]reflect.Type{},
		Proxies:  map[string]reflect.Type{},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}

	// Import template
	imports.Packages["github.com/state-of-the-art/ansiblego/pkg/template"] = imports.Package{
		Binds: map[string]reflect.Value{
			"IsTemplate": reflect.ValueOf(template.IsTemplate),
		},
		Types:    map[string]reflect.Type{},
		Proxies:  map[string]reflect.Type{},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}

	// Import the TaskV1 interface
	imports.Packages["github.com/state-of-the-art/ansiblego/pkg/ansible"] = imports.Package{
		Binds: map[string]reflect.Value{
			"CollectV1":     reflect.ValueOf(CollectV1),
			"TaskV1SetData": reflect.ValueOf(TaskV1SetData),
			"TaskV1GetData": reflect.ValueOf(TaskV1GetData),
			"ModulesList":   reflect.ValueOf(ModulesList),
			"ToYaml":        reflect.ValueOf(ToYaml),
		},
		Types: map[string]reflect.Type{
			"Task":            reflect.TypeOf((*Task)(nil)).Elem(),
			"TaskV1Interface": reflect.TypeOf((*TaskV1Interface)(nil)).Elem(),
			"OrderedMap":      reflect.TypeOf((*OrderedMap)(nil)).Elem(),
			"TAny":            reflect.TypeOf((*TAny)(nil)).Elem(),
			"TAnyMap":         reflect.TypeOf((*TAnyMap)(nil)).Elem(),
			"TAnyList":        reflect.TypeOf((*TAnyList)(nil)).Elem(),
			"TString":         reflect.TypeOf((*TString)(nil)).Elem(),
			"TStringMap":      reflect.TypeOf((*TStringMap)(nil)).Elem(),
			"TStringList":     reflect.TypeOf((*TStringList)(nil)).Elem(),
			"TInt":            reflect.TypeOf((*TInt)(nil)).Elem(),
			"TIntMap":         reflect.TypeOf((*TIntMap)(nil)).Elem(),
			"TIntList":        reflect.TypeOf((*TIntList)(nil)).Elem(),
			"TBool":           reflect.TypeOf((*TBool)(nil)).Elem(),
			"TBoolMap":        reflect.TypeOf((*TBoolMap)(nil)).Elem(),
			"TBoolList":       reflect.TypeOf((*TBoolList)(nil)).Elem(),
		},
		Proxies: map[string]reflect.Type{
			"TaskV1Interface": reflect.TypeOf((*P_TaskV1Interface)(nil)).Elem(),
		},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}
}
