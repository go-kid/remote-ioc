package remote_ioc

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/remote-ioc/defination"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"reflect"
	"time"
)

type iocClient struct {
	c        ClientConfig
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

func (s *iocClient) registerInvoker() {
	metas := s.r.GetComponents(registry.Interface(new(defination.InvokeComponent)))
	s.invokers = lo.SliceToMap(metas, func(m *meta.Meta) (string, *clientComponent) {
		ic := m.Raw.(defination.InvokeComponent)
		serviceId := ic.RemoteServiceId()
		var lb defination.LoadBalancing
		if lbs, isImplLB := m.Raw.(defination.LoadBalancingStrategy); isImplLB {
			lb = lbs.LoadBalancing()
		} else {
			lb = defaultLoadBalancing
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
			httpClient:      resty.New(),
			//debug:           true,
		}
		ic.RegisterInvoker(c.invoke)
		return serviceId, c
	})
}

func (s *iocClient) registerServers() error {
	for _, server := range s.c.Servers {
		var baseUrl = server.Addr + server.RoutePrefix
		var metas = make([]*serverInfo, 0)
		startTime := time.Now()
		_, err := s.client.R().
			SetResult(&metas).
			Get(baseUrl + routeMeta)
		if err != nil {
			return err
		}
		si := &defination.ServerInfo{
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
					serverInfo: []*defination.ServerInfo{si},
				}
			}
		}
	}
	return nil
}

type clientComponent struct {
	m               *meta.Meta
	methodMap       map[string]reflect.Method
	lb              defination.LoadBalancing
	servers         []*defination.ServerInfo
	remoteServiceId string
	httpClient      *resty.Client
	debug           bool
}

func (i *clientComponent) invoke(methodName string, v ...any) ([]any, error) {
	server := i.servers[i.lb(i.servers)]
	method := i.methodMap[methodName]
	body := buildBodyParam(method, v)
	var result = &payload{}
	_, err := i.httpClient.
		SetDebug(i.debug).
		R().
		SetBody(body).
		SetResult(result).
		Post(server.Addr + fmt.Sprintf(routeMethod, i.remoteServiceId, methodName))
	if err != nil {
		return nil, err
	}
	if len(result.Params) != method.Type.NumOut() {
		return nil, errors.New("remote server response parameters not equal")
	}

	var results = make([]any, len(result.Params))
	for i, p := range result.Params {
		value, err := p.ToValue(method.Type.Out(i))
		if err != nil {
			return results, err
		}
		results[i] = value.Interface()
		//results = append(results, value.Interface())
	}
	lenOut := method.Type.NumOut() - 1
	if method.Type.Out(lenOut).String() == "error" && results[lenOut] != nil {
		err = results[lenOut].(error)
	}
	return results, err
}

func defaultLoadBalancing(servers []*defination.ServerInfo) int {
	return 0
}

func buildBodyParam(method reflect.Method, values []any) *payload {
	return &payload{
		Params: lo.Map(values, func(value any, index int) *param {
			i := index + 1
			return &param{
				Order: i,
				Kind:  method.Type.In(i).Kind().String(),
				Value: value,
			}
		}),
	}
}

type serverMeta struct {
	serverInfo []*defination.ServerInfo
	meta       *serverInfo
}
