package dto

import (
	"reflect"
)

type Payload struct {
	Params []*Param `json:"params"`
}

type Param struct {
	Order int    `json:"order"`
	Kind  string `json:"kind"`
	Value any    `json:"value"`
}

func (p *Param) Validate(in reflect.Type) error {
	if in.Kind().String() == p.Kind {
		return nil
	}
	if in.Kind() == reflect.Interface && in.String() == p.Kind {
		return nil
	}
	return &ValidateError{
		Msg:               "invalid parameter type",
		RequiredParamKind: in.Kind().String(),
		RequestParamKind:  p.Kind,
		ParamOrder:        p.Order,
		Value:             p.Value,
	}
}
