package unittest

import (
	"errors"
	"github.com/samber/lo"
	"time"
)

type ServerComponent interface {
	SumI(base, add int) int
	SumS(base, add string) string
	SumF(base, add float64) float64
	And(b1, b2 bool) bool
	SumSliceI(base int, add []int) int
	SumArrayI(base int, add [3]int) int
	SumSliceIV(base int, add ...int) int
	SumObj(obj1, obj2 Obj) Obj
	SumObjPtr(obj1, obj2 *Obj) *Obj
	AddTime(t1 time.Time, duration time.Duration) time.Time
	AddTimePtr(t1 *time.Time, duration *time.Duration) time.Time
	ConvertError(msg string) (string, error)
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

func (s *ServerComponentImpl) AddTime(t1 time.Time, duration time.Duration) time.Time {
	return t1.Add(duration)
}

func (s *ServerComponentImpl) AddTimePtr(t1 *time.Time, duration *time.Duration) time.Time {
	return t1.Add(*duration)
}

func (s *ServerComponentImpl) ConvertError(msg string) (string, error) {
	if msg == "" {
		return "ok", nil
	}
	return "", errors.New(msg)
}
