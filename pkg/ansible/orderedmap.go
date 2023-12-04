package ansible

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type OrderedMap struct {
	data  map[string]any
	order []string
}

func (om *OrderedMap) Keys() (keys []string) {
	keys = make([]string, len(om.order))
	for i, key := range om.order {
		keys[i] = key
	}
	return
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

func (om *OrderedMap) Yaml() (string, error) {
	return ToYaml(om)
}
