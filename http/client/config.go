package client

import (
	"context"
	"fmt"
	"github.com/go-kid/remote-ioc/http/transmission"
	"reflect"
	"time"
)

type Config struct {
	Servers                []ServerConfig
	Debug                  bool
	LoadBalance            LoadBalancing
	SerializationFilters   []SerializationFilter
	DeserializationFilters []DeserializationFilter
}

type ServerConfig struct {
	Addr        string
	RoutePrefix string
}

type ServerInfo struct {
	Addr  string
	Delay time.Duration
}

type LoadBalancing func(servers []*ServerInfo) int

var defaultLoadBalancing = func() LoadBalancing {
	var last int
	return func(servers []*ServerInfo) int {
		last++
		if last == len(servers) {
			last = 0
		}
		return last
	}
}

type SerializationFilter = transmission.SerializationFilter
type DeserializationFilter = transmission.DeserializationFilter

var CtxToMapFilter SerializationFilter = func(inType string, value any) (value2 any, ok bool, err error) {
	ok = inType == "context.Context"
	if ok {
		value2 = getPairs(value.(context.Context))
	}
	return
}

func getPairs(ctx context.Context) map[string]any {
	var m = make(map[string]any)
	for v := reflect.ValueOf(ctx).Elem(); v.Type().String() != "context.emptyCtx"; v = v.FieldByName("Context").Elem().Elem() {
		if v.Type().String() == "context.valueCtx" {
			if v.FieldByName("key").Elem().Kind() == reflect.String {
				key := fmt.Sprintf("%v", v.FieldByName("key"))
				m[key] = fmt.Sprintf("%v", v.FieldByName("val"))
			}
		}
	}
	return m
}
