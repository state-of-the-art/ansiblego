package ansible

// Useful utils for task modules which is available through binding

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

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

	// Processing the known task fields
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

		key := info.Name
		val, ok := fmap.Get(key)
		if !ok {
			// If there is no field name in data keys - check aliases
			for _, alias := range info.Aliases {
				if val, ok = fmap.Get(alias); ok {
					key = alias
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

		if !ok {
			// Skipping not found field
			continue
		}

		rfield := rvalue.Elem().Field(i)

		// Check the field type is ansible T-one
		if !strings.HasPrefix(rfield.Type().Name(), "T") {
			return fmt.Errorf("Unable to process field %q type %q - only T-types should be used", fieldt.Name, rfield.Type().Name())
		}

		/*for i := 0; i < rfield.Interface().Type().NumMethod(); i++ {
			fmt.Println("!!DEBUG:", rfield.Interface().Type().Method(i).Name)
		}*/

		rval := reflect.ValueOf(val)
		rfield.Interface().MethodByName("SetUnknown").Call([]reflect.Value{rval})

		/*if rfield.Kind() != rval.Kind() {
			// Those are not the same types which is alarming, so check if field is an Slice
			if rfield.Kind() == reflect.Slice && rfield.Type().Elem().Kind() == rval.Kind() {
				// Ok field is just an array, so set it's first index to the provided value
				newslice := reflect.MakeSlice(rfield.Type(), 1, 1)
				newslice.Index(0).Set(rval)
				rfield.Set(newslice)
			} else {
				// Unfortunately the types are incompatible, so probably an error in playbook?
				return fmt.Errorf("Unable to set the field `%s` of type %s to value: %q of type %s", key, rfield.Type(), val, rval.Type())
			}
		} else {
			// They could be two slices, but different types of elements - so checking that
			if rfield.Kind() == reflect.Slice && rfield.Type().Elem().Kind() != rval.Type().Elem().Kind() {
				// Aha, they are slices and element types are different, so try to convert
				if !rval.IsNil() && rval.Len() > 0 { // If it's empty - then nothing to set
					if rval.Index(0).Kind() == reflect.Interface && rval.Index(0).Elem().Type() == rfield.Type().Elem() {
						for i := 0; i < rval.Len(); i++ {
							rfield.Set(reflect.Append(rfield, rval.Index(i).Elem()))
						}
					} else {
						return fmt.Errorf("Unable to set the field `%s` of type %s to value: %q of type %s", key, rfield.Type(), val, rval.Type())
					}
				}
			} else {
				// The kinds not slices or their elemenet types are the same - so just set field
				rfield.Set(rval)
			}
		}*/
		// Remove key from fmap to signal that it's processed
		fmap.Pop(key)
	}

	// Check if fmap still contains not processed keys - it's dangerous to not process them aciddentally
	if fmap.Size() > 0 {
		y, err := fmap.Yaml()
		if err != nil {
			y = fmt.Sprintf("Error while encoding the OrderedMap to Yaml: %q, %q", err, fmap)
		}
		return fmt.Errorf("Found next unknown task fields (%d) - maybe not implemented?: %s", fmap.Size(), y)
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
	// Name of the field, if not set then lowercase field.Name will be here
	Name string
	// Aliases for the field to check if no main field is set
	// Max 5 to keep the info comparable to use TString in maps keys
	Aliases [5]string
	// Set the default value if the field is not set
	Default    bool
	DefaultVal any
	// If the field is required to be set no matter what
	Required bool
	// List of available values to set value into, separated by space
	// Max 32 to keep the info comparable to use TString in maps keys
	List [32]string
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
	var aliases []string
	if len(fields) > 1 {
		for _, flag := range fields[1:] {
			kv := strings.SplitN(flag, ":", 2)
			switch kv[0] {
			case "alias":
				aliases = append(aliases, kv[1])
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
				lst := strings.Split(kv[1], " ")
				copy(info.List[:], lst)
			case "req":
				info.Required = true
			default:
				return info, fmt.Errorf("Unsupported flag %q in tag %q of field %s", flag, tag, field.Name)
			}
		}
		tag = fields[0]
	}
	if len(aliases) > 0 {
		copy(info.Aliases[:], aliases)
	}

	if tag != "" {
		info.Name = tag
	} else {
		info.Name = strings.ToLower(field.Name)
	}

	return info, err
}

// Allows to convert any object to good Yaml string
func ToYaml(obj any) (string, error) {
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()
	enc.SetIndent(2)
	if err := enc.Encode(obj); err != nil {
		return "", fmt.Errorf("YAML encode error: %v", err)
	}
	return "---\n" + buf.String(), nil
}
