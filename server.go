package remote_ioc

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/remote-ioc/defination"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"log"
	"reflect"
)

const (
	routeHealth = "/health"
	routeMeta   = "/meta"
	routeMethod = "/component/%s/methods/%s"
)

type iocServer struct {
	c  ServerConfig
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
		g.GET(routeHealth, func(c echo.Context) error {
			return c.JSON(200, map[string]string{
				"status": "ok",
			})
		})
		g.GET(routeMeta, func(c echo.Context) error {
			var metas []*serverInfo
			for _, component := range s.cs {
				metas = append(metas, &serverInfo{
					ServiceId: component.serviceId,
					Methods:   lo.Keys(component.mvm),
				})
			}
			return c.JSON(200, metas)
		})
		for _, component := range s.cs {
			for methodName, method := range component.mvm {
				method := method
				route := fmt.Sprintf(routeMethod, component.serviceId, methodName)
				e.POST(route, func(c echo.Context) error {
					return exportHandler(c, component, method)
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

func exportHandler(c echo.Context, component *serviceComponent, method reflect.Method) error {
	var body = &payload{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	var values = make([]reflect.Value, method.Type.NumIn())
	values[0] = component.m.Value
	for _, p := range body.Params {
		in := method.Type.In(p.Order)
		err = p.Validate(in)
		if err != nil {
			return c.JSON(400, err)
		}
		values[p.Order], err = p.ToValue(in)
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

	var results []*param
	for i := 0; i < method.Type.NumOut(); i++ {
		out := method.Type.Out(i)
		result := convertResult(i, out, resultValues[i])
		results = append(results, result)
	}
	return c.JSON(200, &payload{Params: results})
}

func convertResult(order int, out reflect.Type, result reflect.Value) *param {
	value := result.Interface()
	kind := result.Kind().String()
	if out.String() == "error" {
		kind = out.String()
		if value != nil {
			value = value.(error).Error()
		}
	}
	return &param{
		Order: order,
		Kind:  kind,
		Value: value,
	}
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
}

type serverInfo struct {
	ServiceId string   `json:"service_id"`
	Methods   []string `json:"methods"`
}
