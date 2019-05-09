package model

import (
	"fmt"
	"github.com/araddon/dateparse"
	"strings"
	"time"
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
	case time.Time:
		p.All = append(p.All, DateTimeProperty{Name: name, Value: DateTime(value)})
	case DateTime:
		p.All = append(p.All, DateTimeProperty{Name: name, Value: value})
	case bool:
		p.All = append(p.All, FlagProperty{Name: name, Value: value})
	case int:
		p.All = append(p.All, NumericProperty{Name: name, Value: value})
	default:
		fmt.Printf("Unable to add %q property: %+v type is not known", name, value)
	}
}

func (p Properties) Get(key PropertyName) (interface{}, bool) {
	for index, item := range p.All {
		switch property := item.(type) {
		case TextProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				return property.Value, true
			}
		case DateTimeProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				return property.Value, true
			}
		case FlagProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				return property.Value, true
			}
		case NumericProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				return property.Value, true
			}
		default:
			fmt.Printf("Unable to iterate property %d: type %T is not known", index, item)
		}
	}
	return nil, false
}

func (p Properties) GetDate(key PropertyName) (time.Time, bool) {
	for index, item := range p.All {
		switch property := item.(type) {
		case TextProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				parsed, err := dateparse.ParseAny(property.Value)
				if err != nil {
					fmt.Printf("[%q:%d] Unable to parse date %q: %s\n", property.Name, index, property.Value, err.Error())
					return time.Time{}, false
				}
				return parsed, true
			}
		case DateTimeProperty:
			if strings.EqualFold(string(key), string(property.Name)) {
				return time.Time(property.Value), true
			}
		}
	}
	return time.Time{}, false
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
