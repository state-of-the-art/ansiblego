package modules

// Useful utils for task modules which is available through binding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type OrderedMap struct {
	data  map[string]any
	order []string
}

func (om *OrderedMap) Set(key string, value any) {
	if om.data == nil {
		om.data = make(map[string]any)
	}
	if _, ok := value.(map[string]any); ok {
		// We want to store only OrderedMap instead of map in OrderedMap
		panic("Prohibited to set `map[string]any` as value - use OrderedMap instead")
	}
	om.data[key] = value
	// Check order list and if no key found add it last
	for _, k := range om.order {
		if k == key {
			return
		}
	}
	om.order = append(om.order, key)
}

func (om *OrderedMap) Get(key string) (any, bool) {
	v, ok := om.data[key]
	return v, ok
}

func (om *OrderedMap) Pop(key string) (v any, ok bool) {
	if v, ok = om.data[key]; ok {
		delete(om.data, key)
		limit := len(om.order)
		pos := -1
		for i := 0; i < limit; i++ {
			if om.order[i] == key {
				pos = i
				break
			}
		}
		if pos < 0 {
			// In case pos was not found - seems the implementation error is here
			panic(fmt.Sprintf("No key `%s` found in list: %q", key, om.order))
		}
		om.order = append(om.order[:pos], om.order[pos+1:]...)
	}
	return v, ok
}

func (om *OrderedMap) Size() int {
	return len(om.data)
}

func (om *OrderedMap) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		om.data = make(map[string]any, len(node.Content)/2)
		var key string
		for i, n := range node.Content {
			if i%2 == 0 {
				if err := n.Decode(&key); err != nil {
					return err
				}
				om.order = append(om.order, key)
			} else {
				switch n.Kind {
				case yaml.MappingNode:
					var val OrderedMap
					if err := n.Decode(&val); err != nil {
						return err
					}
					om.data[key] = val
				case yaml.SequenceNode:
					lst := make([]any, len(n.Content))
					for j, a := range n.Content {
						if a.Kind == yaml.MappingNode {
							var val OrderedMap
							if err := a.Decode(&val); err != nil {
								return err
							}
							lst[j] = val
						} else {
							var val any
							if err := a.Decode(&val); err != nil {
								return err
							}
							lst[j] = val
						}
					}
					om.data[key] = lst
				default:
					var val any
					if err := n.Decode(&val); err != nil {
						return err
					}
					om.data[key] = val
				}
			}
		}
	default:
		return fmt.Errorf("Unsupported OrderedMap type: %v", node.Kind)
	}
	return nil
}

func (om *OrderedMap) MarshalYAML() (interface{}, error) {
	nodes := make([]*yaml.Node, len(om.order)*2)
	for i, key := range om.order {
		keyn := &yaml.Node{}
		if err := keyn.Encode(key); err != nil {
			return nil, err
		}
		nodes[i*2] = keyn

		valuen := &yaml.Node{}
		switch val := om.data[key].(type) {
		case OrderedMap:
			d, err := val.MarshalYAML()
			if err != nil {
				return nil, err
			}
			valuen = d.(*yaml.Node)
		case []any:
			valuen.Kind = yaml.SequenceNode
			valuen.Tag = "!!seq"
			lst := make([]*yaml.Node, len(val))
			for j, a := range val {
				switch v := a.(type) {
				case OrderedMap:
					d, err := v.MarshalYAML()
					if err != nil {
						return nil, err
					}
					lst[j] = d.(*yaml.Node)
				default:
					d := &yaml.Node{}
					if err := d.Encode(v); err != nil {
						return nil, err
					}
					lst[j] = d
				}
			}
			valuen.Content = lst
		default:
			if err := valuen.Encode(val); err != nil {
				return nil, err
			}
		}
		nodes[i*2+1] = valuen
	}
	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: nodes,
	}, nil
}

// TaskV1 struct field tags: `task:"[NAME][,OPTS]"`
//   * NAME - the yaml key of the field, if `-` - will be skipped
//   * OPTS - various options, comma separated:
//     * def:VALUE - default value of the field
//     * alias:ALIAS - alias for the key, will be processed
//     * list:VAL[ ...] - list to choose the value from
//     * req - required field

// To set data into the typical task struct according to defined field tags
func TaskV1SetData(task_ptr any, fmap OrderedMap) error {
	rtype := reflect.TypeOf(task_ptr)
	if rtype.Kind() != reflect.Ptr {
		return fmt.Errorf("Can not process non-pointer task: %s", rtype.Kind())
	}
	rtype = rtype.Elem()
	if rtype.Kind() != reflect.Struct {
		return fmt.Errorf("Unable to set data for non-struct: %s", rtype.Kind())
	}

	// Processing the fields
	rvalue := reflect.ValueOf(task_ptr)
	for i := 0; i < rtype.NumField(); i++ {
		fieldt := rtype.Field(i)
		info, err := TaskV1FieldInfo(&fieldt)
		if err != nil {
			return err
		}
		if info.Skip {
			continue
		}

		val, ok := fmap.Get(info.Name)
		if !ok {
			// If there is no field name in data keys - check aliases
			for _, alias := range info.Aliases {
				if val, ok = fmap.Get(alias); ok {
					break
				}
			}
			if !ok {
				if info.Default {
					// If value is not set and we have default - let it be
					val = info.DefaultVal
					ok = true
				} else if info.Required {
					// If value is required and not set - fail
					return fmt.Errorf("Unable to find the required value for field `%s`", fieldt.Name)
				}
			}
		}
		if len(info.List) > 0 {
			// Check if the value in the list
			found := false
			for _, v := range info.List {
				if val == v {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("The value of field `%s` is not in the defined list %q", fieldt.Name, info.List)
			}
		}
		if ok {
			rvalue.Elem().Field(i).Set(reflect.ValueOf(val))
		}
	}
	return nil
}

// To get data from a typical task struct according to defined field tags
// It returns list of key-value pairs to preserve the order of fields
func TaskV1GetData(task_ptr any) (fmap OrderedMap) {
	rval := reflect.ValueOf(task_ptr).Elem()
	rtype := reflect.TypeOf(task_ptr).Elem()

	for i := 0; i < rtype.NumField(); i++ {
		fieldt := rtype.Field(i)
		info, err := TaskV1FieldInfo(&fieldt)
		if err != nil || info.Skip {
			continue
		}
		if info.Default {
			if info.DefaultVal == rval.Field(i).Interface() {
				continue
			}
		} else {
			// Check if the value was set or not
			fieldv := rval.Field(i)
			switch fieldt.Type.Kind() {
			case reflect.Bool, reflect.Int, reflect.String, reflect.Struct:
				if fieldv.IsZero() {
					continue
				}
			case reflect.Array, reflect.Map, reflect.Slice:
				if fieldv.IsNil() || fieldv.Len() == 0 {
					continue
				}
			}
		}
		fmap.Set(strings.ToLower(fieldt.Name), rval.Field(i).Interface())
	}

	return fmap
}

type fieldInfo struct {
	Skip bool

	Name string // Name of the field, if not set then lowercase field.Name will be here

	Aliases []string // Aliases for the field to check if no main field is set

	Default    bool // Set the default value if the field is not set
	DefaultVal any

	Required bool // If the field is required to be set no matter what

	List []string // Only listed items are available to be set as value
}

// Get information from the TaskV1 field tags, useful when processing semi-automatically
func TaskV1FieldInfo(field *reflect.StructField) (info fieldInfo, err error) {
	// For some reason field.PkgPath is empty for structs received from gomacro script
	// so using stupid comparison of first symbol in name to lowercase of it as last resort
	if field.PkgPath != "" || field.Anonymous || field.Name[0] == strings.ToLower(field.Name)[0] {
		// Skip private or embedded field
		info.Skip = true
		return
	}

	tag := field.Tag.Get("task")
	if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
		tag = string(field.Tag)
	}
	if tag == "-" {
		info.Skip = true
		return // No need to process
	}

	fields := strings.Split(tag, ",")
	if len(fields) > 1 {
		for _, flag := range fields[1:] {
			kv := strings.SplitN(flag, ":", 2)
			switch kv[0] {
			case "alias":
				info.Aliases = append(info.Aliases, kv[1])
			case "def":
				info.Default = true
				switch field.Type.Kind() {
				case reflect.Bool:
					if info.DefaultVal, err = strconv.ParseBool(kv[1]); err != nil {
						return info, fmt.Errorf("Incorrect field `%s` bool value '%s': %s", field.Name, flag, err)
					}
				case reflect.Int:
					if info.DefaultVal, err = strconv.Atoi(kv[1]); err != nil {
						return info, fmt.Errorf("Incorrect field `%s` int value '%s': %s", field.Name, flag, err)
					}
				case reflect.String:
					info.DefaultVal = kv[1]
				default:
					return info, fmt.Errorf("Unsupported default value for field `%s` type '%s'", field.Name, field.Type.Kind())
				}
			case "list":
				info.List = strings.Split(kv[1], " ")
			case "req":
				info.Required = true
			default:
				return info, fmt.Errorf("Unsupported flag %q in tag %q of field %s", flag, tag, field.Name)
			}
		}
		tag = fields[0]
	}

	if tag != "" {
		info.Name = tag
	} else {
		info.Name = strings.ToLower(field.Name)
	}

	return info, err
}
