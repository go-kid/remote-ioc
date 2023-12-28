package client

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/remote-ioc/defination"
	"github.com/go-kid/remote-ioc/http/constant"
	"github.com/go-kid/remote-ioc/http/dto"
	"github.com/go-kid/remote-ioc/http/transmission"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"reflect"
	"time"
)

type iocClient struct {
	c        Config
	r        registry.Registry
	servers  map[string]*serverMeta
	invokers map[string]*clientComponent

	client *resty.Client
}

func (s *iocClient) Init() error {
	s.servers = make(map[string]*serverMeta)
	s.client = resty.New()

	err := s.registerServers()
	if err != nil {
		return err
	}
	s.registerInvoker()

	return nil
}

func (s *iocClient) registerServers() error {
	for _, server := range s.c.Servers {
		var baseUrl = server.Addr + server.RoutePrefix
		var metas = make([]*dto.ServerInfo, 0)
		startTime := time.Now()
		_, err := s.client.R().
			SetResult(&metas).
			Get(baseUrl + constant.RouteMeta)
		if err != nil {
			return err
		}
		si := &ServerInfo{
			Addr:  baseUrl,
			Delay: time.Now().Sub(startTime),
		}
		for _, info := range metas {
			if sm, ok := s.servers[info.ServiceId]; ok {
				if !reflect.DeepEqual(sm.meta, info) {
					return fmt.Errorf("remote component %s not equal in multi-server", info.ServiceId)
				}
				sm.serverInfo = append(sm.serverInfo, si)
			} else {
				s.servers[info.ServiceId] = &serverMeta{
					meta:       info,
					serverInfo: []*ServerInfo{si},
				}
			}
		}
	}
	return nil
}

func (s *iocClient) registerInvoker() {
	metas := s.r.GetComponents(registry.Interface(new(defination.InvokeComponent)))
	s.invokers = lo.SliceToMap(metas, func(m *meta.Meta) (string, *clientComponent) {
		ic := m.Raw.(defination.InvokeComponent)
		serviceId := ic.RemoteServiceId()
		var lb = s.c.LoadBalance
		if lb == nil {
			lb = defaultLoadBalancing()
		}
		sm := s.servers[serviceId]
		var methodMap = make(map[string]reflect.Method)
		for _, methodName := range sm.meta.Methods {
			if method, ok := m.Type.MethodByName(methodName); ok {
				methodMap[methodName] = method
			} else {
				panic(fmt.Errorf("remote component %s method %s not found", sm.meta.ServiceId, methodName))
			}
		}
		c := &clientComponent{
			m:               m,
			methodMap:       methodMap,
			lb:              lb,
			servers:         sm.serverInfo,
			remoteServiceId: serviceId,
			httpClient:      resty.New().SetDebug(s.c.Debug),
			sFilters:        s.c.SerializationFilters,
			dsFilters:       s.c.DeserializationFilters,
		}
		ic.RegisterInvoker(c.invoke)
		return serviceId, c
	})
}

type clientComponent struct {
	m               *meta.Meta
	methodMap       map[string]reflect.Method
	lb              LoadBalancing
	servers         []*ServerInfo
	remoteServiceId string
	httpClient      *resty.Client
	sFilters        []SerializationFilter
	dsFilters       []DeserializationFilter
}

func (i *clientComponent) invoke(methodName string, v ...any) (results []any, err error) {
	server := i.servers[i.lb(i.servers)]
	method := i.methodMap[methodName]
	results = make([]any, method.Type.NumOut())
	for i := 0; i < method.Type.NumOut(); i++ {
		results[i] = reflect.New(method.Type.Out(i)).Elem().Interface()
	}

	var body *dto.Payload
	body, err = i.buildBodyParam(method, v)
	if err != nil {
		return
	}
	var resp = &dto.Payload{}
	start := time.Now()
	_, err = i.httpClient.
		R().
		SetBody(body).
		SetResult(resp).
		Post(server.Addr + fmt.Sprintf(constant.RouteMethod, i.remoteServiceId, methodName))
	if err != nil {
		return
	}
	server.Delay = time.Now().Sub(start)

	if len(resp.Params) != method.Type.NumOut() {
		err = errors.New("remote server response parameters not equal")
		return
	}

	for index, p := range resp.Params {
		var value reflect.Value
		value, err = transmission.DecryptParam(p, method.Type.Out(index), i.dsFilters)
		if err != nil {
			return
		}

		results[index] = value.Interface()
	}
	return
}

func (i *clientComponent) buildBodyParam(method reflect.Method, values []any) (*dto.Payload, error) {
	var params []*dto.Param
	for index := 1; index < method.Type.NumIn(); index++ {
		param, err := transmission.EncryptParam(index, method.Type.In(index), values[index-1], i.sFilters)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	return &dto.Payload{
		Params: params,
	}, nil
}

type serverMeta struct {
	serverInfo []*ServerInfo
	meta       *dto.ServerInfo
}
