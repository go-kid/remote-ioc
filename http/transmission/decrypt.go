package transmission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kid/ioc/util/fas"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/remote-ioc/http/dto"
	"reflect"
)

type DeserializationFilter func(p *dto.Param, inType reflect.Type) (reflect.Value, bool, error)

func DecryptParam(p *dto.Param, in reflect.Type, filters []DeserializationFilter) (value reflect.Value, err error) {
	filters = append(filters, defaultDeserializationFilters...)
	if in.Kind() == reflect.Interface {
		var find bool
		for _, filter := range filters {
			value, find, err = filter(p, in)
			if err != nil {
				return value, err
			}
			if find {
				break
			}
		}
		if !find {
			err = fmt.Errorf("decrypt interface \"%s\" faild: not found supported deserialization filter", in.Kind())
		}
	} else {
		value, err = convertJsonValue(in, p.Value)
	}
	if err != nil {
		err = &dto.ConvertError{
			Param: p,
			Err:   fmt.Sprintf("parameter[%d]%s", p.Order, err),
		}
	}
	return value, err
}

var defaultDeserializationFilters = []DeserializationFilter{
	defaultContextDeserializationFilter,
	defaultErrorDeserializationFilter,
}

var defaultContextDeserializationFilter DeserializationFilter = func(p *dto.Param, inType reflect.Type) (val reflect.Value, ok bool, err error) {
	ok = p.Kind == "context.Context"
	if ok {
		val = reflect.ValueOf(context.Background())
		ok = true
	}
	return
}

var defaultErrorDeserializationFilter DeserializationFilter = func(p *dto.Param, inType reflect.Type) (val reflect.Value, ok bool, err error) {
	ok = p.Kind == "error"
	if !ok {
		return
	}
	if p.Value == nil {
		val = reflect.New(inType).Elem()
	} else if s, ok := p.Value.(string); ok {
		//val = reflect.ValueOf(errors.New(s))
		err = errors.New(s)
	} else {
		err = errors.New(": value is not a string")
	}
	return
}

func convertJsonValue(in reflect.Type, val any) (value reflect.Value, err error) {
	value = reflectx.New(in)
	if val == nil {
		value = fas.TernaryOp(in.Kind() == reflect.Pointer, value, value.Elem())
		return
	}
	if msl, ok := value.Interface().(json.Unmarshaler); ok {
		err = msl.UnmarshalJSON([]byte(fmt.Sprintf("%#v", val)))
		if err != nil {
			return
		}
		value = fas.TernaryOp(in.Kind() == reflect.Pointer, value, value.Elem())
		return
	}

	switch in.Kind() {
	case reflect.String:
		if s, ok := val.(string); ok {
			value = value.Elem()
			value.SetString(s)
		} else {
			err = errors.New(": value is not a string")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if f, ok := val.(float64); ok {
			value = value.Elem()
			value.SetInt(int64(f))
		} else {
			err = errors.New(": value is not a number")
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := val.(float64); ok {
			value = value.Elem()
			value.SetFloat(f)
		} else {
			err = errors.New(": value is not a number")
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			value = value.Elem()
			value.SetBool(b)
		} else {
			err = errors.New(": value is not a boolean")
		}
	case reflect.Struct:
		if vm, ok := val.(map[string]any); ok {
			value = value.Elem()
			err = reflectx.ForEachFieldV2(in, value, true, func(field reflect.StructField, value reflect.Value) error {
				key := field.Tag.Get("json")
				v, err := convertJsonValue(field.Type, vm[key])
				if err != nil {
					return fmt.Errorf(".%s%v", key, err)
				}
				value.Set(v)
				return nil
			})
		} else {
			err = errors.New(": value is not a object")
		}
	case reflect.Array:
		if anies, ok := val.([]any); ok {
			value = value.Elem()
			nt := in.Elem()
			if len(anies) > value.Len() {
				err = errors.New(": array index out of range")
				return
			}
			for i, item := range anies {
				var v reflect.Value
				v, err = convertJsonValue(nt, item)
				if err != nil {
					err = fmt.Errorf(".[%d]%s.$%d%v", value.Len(), nt.String(), i+1, err)
					break
				}
				value.Index(i).Set(v)
			}
		} else {
			err = errors.New(": value is not an array")
		}
	case reflect.Slice:
		if anies, ok := val.([]any); ok {
			value = value.Elem()
			nt := in.Elem()
			var values []reflect.Value
			for i, item := range anies {
				var v reflect.Value
				v, err = convertJsonValue(nt, item)
				if err != nil {
					err = fmt.Errorf(".[]%s.$%d%v", nt.String(), i+1, err)
					break
				}
				values = append(values, v)
			}
			value.Set(reflect.Append(value, values...))
		} else {
			err = errors.New(": value is not an array")
		}

	case reflect.Pointer:
		var v reflect.Value
		v, err = convertJsonValue(in.Elem(), val)
		if err != nil {
			return
		}
		value.Elem().Set(v)
	case reflect.Interface: //only handle the interface of error
		if s, ok := val.(string); ok {
			value = value.Elem()
			value.Set(reflect.ValueOf(errors.New(s)))
		} else {
			err = errors.New(": value is not a string")
		}
	default:
		err = errors.New(": unsupported type")
	}
	return
}
