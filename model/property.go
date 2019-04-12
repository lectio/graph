package model

import (
	"fmt"
)

type PropertyName string

func MakeProperties() *Properties {
	result := new(Properties)
	return result
}

func (p *Properties) add(name PropertyName, v interface{}) {
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
