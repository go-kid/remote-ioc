package dto

import "encoding/json"

type ValidateError struct {
	Msg               string `json:"msg"`
	RequiredParamKind string `json:"required_param_kind"`
	RequestParamKind  string `json:"request_param_kind"`
	ParamOrder        int    `json:"param_order"`
	Value             any    `json:"value"`
}

func (e *ValidateError) Error() string {
	marshal, _ := json.Marshal(e)
	return string(marshal)
}

type ConvertError struct {
	*Param `json:",inline"`
	Err    string `json:"error"`
}

func (e *ConvertError) Error() string {
	marshal, _ := json.Marshal(e)
	return string(marshal)
}
