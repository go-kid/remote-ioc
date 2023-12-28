package http

import (
	"context"
	"github.com/go-kid/remote-ioc/defination"
	"time"
)

type ServerComponentInvoker struct {
	Invoke defination.Invoke
}

func (s *ServerComponentInvoker) RegisterInvoker(invoke defination.Invoke) {
	s.Invoke = invoke
}

func (s *ServerComponentInvoker) RemoteServiceId() string {
	return "MathServer"
}

func (s *ServerComponentInvoker) SumI(base, add int) int {
	anies, err := s.Invoke("SumI", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(int)
}

func (s *ServerComponentInvoker) SumS(base, add string) string {
	anies, err := s.Invoke("SumS", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(string)
}

func (s *ServerComponentInvoker) SumF(base, add float64) float64 {
	anies, err := s.Invoke("SumF", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(float64)
}

func (s *ServerComponentInvoker) And(b1, b2 bool) bool {
	anies, err := s.Invoke("And", b1, b2)
	if err != nil {
		panic(err)
	}
	return anies[0].(bool)
}

func (s *ServerComponentInvoker) SumSliceI(base int, add []int) int {
	anies, err := s.Invoke("SumSliceI", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(int)
}

func (s *ServerComponentInvoker) SumArrayI(base int, add [3]int) int {
	anies, err := s.Invoke("SumArrayI", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(int)
}

func (s *ServerComponentInvoker) SumSliceIV(base int, add ...int) int {
	anies, err := s.Invoke("SumSliceIV", base, add)
	if err != nil {
		panic(err)
	}
	return anies[0].(int)
}

func (s *ServerComponentInvoker) SumObj(obj1, obj2 Obj) Obj {
	anies, err := s.Invoke("SumObj", obj1, obj2)
	if err != nil {
		panic(err)
	}
	return anies[0].(Obj)
}

func (s *ServerComponentInvoker) SumObjPtr(obj1, obj2 *Obj) *Obj {
	anies, err := s.Invoke("SumObjPtr", obj1, obj2)
	if err != nil {
		panic(err)
	}
	return anies[0].(*Obj)
}

func (s *ServerComponentInvoker) AddTime(t1 time.Time, duration time.Duration) time.Time {
	anies, err := s.Invoke("AddTime", t1, duration)
	if err != nil {
		panic(err)
	}
	return anies[0].(time.Time)
}

func (s *ServerComponentInvoker) AddTimePtr(t1 *time.Time, duration *time.Duration) time.Time {
	anies, err := s.Invoke("AddTimePtr", t1, duration)
	if err != nil {
		panic(err)
	}
	return anies[0].(time.Time)
}

func (s *ServerComponentInvoker) ConvertError(msg string) (string, error) {
	anies, err := s.Invoke("ConvertError", msg)
	return anies[0].(string), err
}

func (s *ServerComponentInvoker) WithContext(ctx context.Context) string {
	anies, err := s.Invoke("WithContext", ctx)
	if err != nil {
		panic(err)
	}
	return anies[0].(string)
}
