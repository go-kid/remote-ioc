package transmission

import (
	"fmt"
	"github.com/go-kid/remote-ioc/http/dto"
	"reflect"
)

func EncryptParam(order int, paramType reflect.Type, value any, filters []SerializationFilter) (*dto.Param, error) {
	var kind = paramType.Kind().String()
	filters = append(filters, defaultSerializationFilters...)
	if paramType.Kind() == reflect.Interface {
		kind = paramType.String()
		var find bool
		for _, filter := range filters {
			val2, ok, err := filter(kind, value)
			if err != nil {
				return nil, err
			}
			if ok {
				value = val2
				find = true
				break
			}
		}
		if !find {
			return nil, fmt.Errorf("encrypt interface \"%s\" faild: not found supported serialization filter", kind)
		}
	}
	return &dto.Param{
		Order: order,
		Kind:  kind,
		Value: value,
	}, nil
}

type SerializationFilter func(inType string, value any) (val2 any, ok bool, err error)

var defaultSerializationFilters = []SerializationFilter{
	defaultContextSerializationFilter,
	defaultErrorSerializationFilter,
}

var defaultContextSerializationFilter SerializationFilter = func(inType string, value any) (val2 any, ok bool, err error) {
	ok = inType == "context.Context"
	if ok {
		val2 = nil
	}
	return
}

var defaultErrorSerializationFilter SerializationFilter = func(inType string, value any) (val2 any, ok bool, err error) {
	ok = inType == "error"
	if !ok {
		return
	}
	if value != nil {
		val2 = value.(error).Error()
	}
	return
}
