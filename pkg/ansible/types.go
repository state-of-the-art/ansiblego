package ansible

// Contains dual types which can contain template or golang type value

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/template"
)

type TAny struct {
	template string
	value    any

	// Used to store field metadata to verify value
	meta fieldInfo
}
type TAnyMap map[TString]TAny
type TAnyList []TAny

type TString struct {
	TAny
}
type TStringMap map[TString]TString
type TStringList []TString

type TInt struct {
	TAny
}
type TIntMap map[TString]TInt
type TIntList []TInt

type TBool struct {
	TAny
}
type TBoolMap map[TString]TBool
type TBoolList []TBool

func NewTString(val string) (out TString) {
	out.value = val
	return out
}

func NewTInt(val int) (out TInt) {
	out.value = val
	return out
}

func NewTBool(val bool) (out TBool) {
	out.value = val
	return out
}

func (t *TAny) UnmarshalYAML(val *yaml.Node) error {
	// Check if it's template
	if val.Kind == yaml.ScalarNode && val.Tag == "!!str" && template.IsTemplate(val.Value) {
		t.template = val.Value
		t.value = nil
		return nil
	}
	return val.Decode(&t.value)
}
func (t *TAny) MarshalYAML() (any, error) {
	node := &yaml.Node{}

	// If it's a template value - return template, otherwise static value
	if t.template != "" {
		if err := node.Encode(t.template); err != nil {
			return nil, fmt.Errorf("Unable to encode template data: %v", err)
		}
	} else {
		if err := node.Encode(t.value); err != nil {
			return nil, fmt.Errorf("Unable to encode value data: %v", err)
		}
	}

	return node, nil
}

func (t *TAny) IsEmpty() bool {
	val, is_string := t.value.(string)
	return (t.value == nil || is_string && val == "") && t.template == ""
}

func (t TAny) String() string {
	if t.template != "" {
		return t.template
	}
	return fmt.Sprintf("%v", t.value)
}

func (t *TAny) SetTemplate(tmpl string) {
	t.template = tmpl
}

func (t *TAny) SetMeta(mt fieldInfo) {
	t.meta = mt
}

func (t *TAny) SetUnknown(val any) {
	if v, is_string := val.(string); is_string && template.IsTemplate(v) {
		t.template = v
		return
	}
	t.value = val
}

func (t *TString) SetValue(val string) {
	t.value = val
}

func (t *TInt) SetValue(val int) {
	t.value = val
}

func (t *TBool) SetValue(val bool) {
	t.value = val
}

func (t *TString) Val() string {
	if t.value == nil {
		if t.meta.Default {
			return t.meta.DefaultVal.(string)
		}
		return ""
	}
	return t.value.(string)
}

func (t *TStringList) Val() (out []string) {
	for _, v := range *t {
		out = append(out, v.Val())
	}

	return out
}

func (t *TInt) Val() int {
	if t.value == nil {
		if t.meta.Default {
			return t.meta.DefaultVal.(int)
		}
		return 0
	}
	return t.value.(int)
}

func (t *TBool) Val() bool {
	if t.value == nil {
		if t.meta.Default {
			return t.meta.DefaultVal.(bool)
		}
		return false
	}
	return t.value.(bool)
}

// Use metadata to verify the stored value
func (t *TAny) Validate() error {
	if t.meta.Required && t.IsEmpty() {
		return fmt.Errorf("Unable to find the required value for field %q", t.meta.Name)
	}
	if t.value == nil && t.template != "" {
		// Value is not yet received from the template
		return nil
	}
	if len(t.meta.List) > 0 {
		// Check if the value in the list
		found := false
		for _, v := range t.meta.List {
			if t.value == v {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("The value of field %q is not in the defined list %q", t.meta.Name, t.meta.List)
		}
	}
	return nil
}
