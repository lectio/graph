package model

import (
	"fmt"
)

type PropertyName string

func MakeProperties() *Properties {
	result := new(Properties)
	return result
}

func (p *Properties) Add(name PropertyName, v interface{}) {
	switch value := v.(type) {
	case string:
		p.All = append(p.All, TextProperty{Name: name, Value: value})
	case bool:
		p.All = append(p.All, FlagProperty{Name: name, Value: value})
	case int:
		p.All = append(p.All, NumericProperty{Name: name, Value: value})
	default:
		fmt.Printf("Unable to add %q property: %+v type is not known", name, value)
	}
}

func (p Properties) ForEach(do func(key PropertyName, value interface{})) {
	for index, item := range p.All {
		switch property := item.(type) {
		case TextProperty:
			do(property.Name, property.Value)
		case FlagProperty:
			do(property.Name, property.Value)
		case NumericProperty:
			do(property.Name, property.Value)
		default:
			fmt.Printf("Unable to iterate property %d: type %T is not known", index, item)
		}
	}
}
