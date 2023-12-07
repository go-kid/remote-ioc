package unittest

import (
	"errors"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	remote_ioc "github.com/go-kid/remote-ioc"
	"github.com/samber/lo"
	"testing"
	"time"
)

type ServerComponent interface {
	Sum(base int, add ...int) int
}

type ServerComponentImpl struct {
}

func (s *ServerComponentImpl) RemoteServiceId() string {
	return "MathServer"
}

//func (s *ServerComponentImpl) ExportMethods() []string {
//	return []string{
//		"SumI",
//		"SumS",
//		"SumF",
//		"And",
//		"SumSliceI",
//		"SumArrayI",
//		"SumSliceIV",
//		"SumObj",
//		"SumObjPtr",
//		"SumTime",
//		"SumTimePtr",
//		"AddTime",
//		"ConvertError",
//	}
//}

//func (s *ServerComponentImpl) ExcludeMethods() []string {
//	return []string{
//		"SumI",
//	}
//}

func (s *ServerComponentImpl) ReplaceMethods() map[string]string {
	return map[string]string{
		"SumI": "SumInt",
	}
}

func (s *ServerComponentImpl) SumI(base, add int) int {
	return base + add
}

func (s *ServerComponentImpl) SumS(base, add string) string {
	return base + add
}

func (s *ServerComponentImpl) SumF(base, add float64) float64 {
	return base + add
}

func (s *ServerComponentImpl) And(b1, b2 bool) bool {
	return b1 && b2
}

func (s *ServerComponentImpl) SumSliceI(base int, add []int) int {
	var result = base
	for _, i := range add {
		result += i
	}
	return result
}

func (s *ServerComponentImpl) SumArrayI(base int, add [3]int) int {
	var result = base
	for _, i := range add {
		result += i
	}
	return result
}

func (s *ServerComponentImpl) SumSliceIV(base int, add ...int) int {
	var result = base
	for _, i := range add {
		result += i
	}
	return result
}

type Obj struct {
	Int    int    `json:"int"`
	String string `json:"string"`
	Subs   []*Sub `json:"subs"`
}

type Sub struct {
	Float float64 `json:"float"`
}

func (s *ServerComponentImpl) SumObj(obj1, obj2 Obj) Obj {
	return *s.SumObjPtr(&obj1, &obj2)
}

func (s *ServerComponentImpl) SumObjPtr(obj1, obj2 *Obj) *Obj {
	lo.SumBy(obj1.Subs, func(item *Sub) float64 {
		return item.Float
	})
	return &Obj{
		Int:    s.SumI(obj1.Int, obj2.Int),
		String: s.SumS(obj1.String, obj2.String),
		Subs: []*Sub{
			{
				Float: lo.SumBy(obj1.Subs, func(item *Sub) float64 {
					return item.Float
				}) + lo.SumBy(obj2.Subs, func(item *Sub) float64 {
					return item.Float
				}),
			},
		},
	}
}

func (s *ServerComponentImpl) SumTime(t1, t2 time.Time) time.Time {
	return t1.AddDate(t2.Year(), int(t2.Month()), t2.Day())
}

func (s *ServerComponentImpl) SumTimePtr(t1, t2 *time.Time) time.Time {
	return t1.AddDate(t2.Year(), int(t2.Month()), t2.Day())
}

func (s *ServerComponentImpl) AddTime(t1 time.Time, duration time.Duration) time.Time {
	return t1.Add(duration)
}

func (s *ServerComponentImpl) ConvertError(msg string) (string, error) {
	if msg == "" {
		return "ok", nil
	}
	return "", errors.New(msg)
}

type ClientApp struct {
	C ServerComponent `remote:""`
}

func TestRemoteIOC(t *testing.T) {
	var s = &ServerComponentImpl{}
	ioc.RunTest(t,
		app.SetComponents(s),
		remote_ioc.Handle(remote_ioc.ServerConfig{
			Addr:        ":8888",
			RoutePrefix: "",
		}),
	)
	time.Sleep(time.Hour)
}
