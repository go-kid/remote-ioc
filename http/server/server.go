package server

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/remote-ioc/defination"
	"github.com/go-kid/remote-ioc/http/constant"
	"github.com/go-kid/remote-ioc/http/dto"
	"github.com/go-kid/remote-ioc/http/transmission"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"log"
	"reflect"
	"sort"
)

type iocServer struct {
	c  Config
	r  registry.Registry
	cs []*serviceComponent
}

func (s *iocServer) Order() int {
	return 999
}

func (s *iocServer) Run() error {
	s.registerRemoteHandler()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	{
		g := e.Group(s.c.RoutePrefix)
		g.GET(constant.RouteHealth, func(c echo.Context) error {
			return c.JSON(200, map[string]string{
				"status": "ok",
			})
		})
		g.GET(constant.RouteMeta, func(c echo.Context) error {
			var metas []*dto.ServerInfo
			for _, component := range s.cs {
				keys := lo.Keys(component.mvm)
				sort.Slice(keys, func(i, j int) bool {
					return keys[i] < keys[j]
				})
				metas = append(metas, &dto.ServerInfo{
					ServiceId: component.serviceId,
					Methods:   keys,
				})
			}
			return c.JSON(200, metas)
		})
		for _, component := range s.cs {
			for methodName, method := range component.mvm {
				method := method
				route := fmt.Sprintf(constant.RouteMethod, component.serviceId, methodName)
				e.POST(route, func(c echo.Context) error {
					return component.exportHandler(c, method)
				})
			}
		}
	}

	go func() {
		log.Printf("[remote-ioc] remote component started on: %s", s.c.Addr)
		e.Logger.Fatal(
			e.Start(s.c.Addr),
		)
	}()
	return nil
}

func (s *iocServer) registerRemoteHandler() {
	metas := s.r.GetComponents(registry.Interface(new(defination.RemoteComponent)))
	s.cs = lo.Map(metas, func(m *meta.Meta, index int) *serviceComponent {
		var methods []string
		if export, ok := m.Raw.(defination.RemoteMethodExport); ok {
			methods = export.ExportMethods()
			methods = lo.Without(methods, "ExportMethods")
		} else {
			for i := 0; i < m.Type.NumMethod(); i++ {
				if mi := m.Type.Method(i); mi.IsExported() {
					methods = append(methods, mi.Name)
				}
			}
		}

		if exclude, ok := m.Raw.(defination.RemoteMethodExclude); ok {
			methods = lo.Without(methods, exclude.ExcludeMethods()...)
			methods = lo.Without(methods, "ExcludeMethods")
		}

		methodMap := lo.SliceToMap(methods, func(item string) (string, reflect.Method) {
			method, ok := m.Type.MethodByName(item)
			if !ok {
				panic("invalid export method: " + item)
			}
			return item, method
		})

		return &serviceComponent{
			m:         m,
			serviceId: m.Raw.(defination.RemoteComponent).RemoteServiceId(),
			mvm:       methodMap,
		}
	})
}

type serviceComponent struct {
	m         *meta.Meta
	serviceId string
	mvm       map[string]reflect.Method
	sFilters  []SerializationFilter
	dsFilters []DeserializationFilter
}

func (s *serviceComponent) exportHandler(c echo.Context, method reflect.Method) error {
	var body = &dto.Payload{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	var values = make([]reflect.Value, method.Type.NumIn())
	values[0] = s.m.Value
	for _, p := range body.Params {
		in := method.Type.In(p.Order)
		err = p.Validate(in)
		if err != nil {
			return c.JSON(400, err)
		}
		values[p.Order], err = transmission.DecryptParam(p, in, s.dsFilters)
		if err != nil {
			return c.JSON(400, err)
		}
	}
	var resultValues []reflect.Value
	if method.Type.IsVariadic() {
		resultValues = method.Func.CallSlice(values)
	} else {
		resultValues = method.Func.Call(values)
	}
	payload, err := s.buildResponseParam(method, resultValues)
	if err != nil {
		return c.JSON(400, err)
	}

	return c.JSON(200, payload)
}

func (s *serviceComponent) buildResponseParam(method reflect.Method, values []reflect.Value) (*dto.Payload, error) {
	var params []*dto.Param
	for i := 0; i < method.Type.NumOut(); i++ {
		out := method.Type.Out(i)
		param, err := transmission.EncryptParam(i+1, out, values[i].Interface(), s.sFilters)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
		//result := convertResult(i+1, out, values[i])
		//params = append(params, result)
	}
	return &dto.Payload{Params: params}, nil
}
