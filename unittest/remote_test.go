package unittest

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	remote_ioc "github.com/go-kid/remote-ioc"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRemoteIOC(t *testing.T) {
	var s = &ServerComponentImpl{}
	ioc.RunTest(t,
		app.SetComponents(s),
		remote_ioc.Handle(remote_ioc.ServerConfig{
			Addr:        ":8888",
			RoutePrefix: "",
		}),
	)

	var c = &ClientApp{}
	ioc.RunTest(t,
		app.SetComponents(c, &ServerComponentInvoker{}),
		remote_ioc.Remote(remote_ioc.ClientConfig{
			Servers: []remote_ioc.ServerConfig{
				{
					Addr:        "http://localhost:8888",
					RoutePrefix: "",
				},
			},
		}),
	)
	t.Run("SumI", func(t *testing.T) {
		assert.Equal(t, c.C.SumI(1, 2), s.SumI(1, 2))
	})
	t.Run("SumS", func(t *testing.T) {
		assert.Equal(t, c.C.SumS("a", "b"), s.SumS("a", "b"))
	})
	t.Run("SumF", func(t *testing.T) {
		assert.Equal(t, c.C.SumF(1.2, 3.4), s.SumF(1.2, 3.4))
	})
	t.Run("And", func(t *testing.T) {
		assert.Equal(t, c.C.And(true, false), s.And(true, false))
	})
	t.Run("SumSliceI", func(t *testing.T) {
		assert.Equal(t, c.C.SumSliceI(1, []int{2, 3}), s.SumSliceI(1, []int{2, 3}))
	})
	t.Run("SumArrayI", func(t *testing.T) {
		assert.Equal(t, c.C.SumArrayI(1, [3]int{2, 3, 4}), s.SumArrayI(1, [3]int{2, 3, 4}))
	})
	t.Run("SumSliceIV", func(t *testing.T) {
		assert.Equal(t, c.C.SumSliceIV(1, 2, 3), s.SumSliceIV(1, 2, 3))
	})
	obj1 := Obj{
		Int:    1,
		String: "a",
		Subs:   []*Sub{},
	}
	obj2 := Obj{
		Int:    2,
		String: "b",
		Subs:   []*Sub{},
	}
	t.Run("SumObj", func(t *testing.T) {
		assert.Equal(t, c.C.SumObj(obj1, obj2), s.SumObj(obj1, obj2))
	})
	t.Run("SumObjPtr", func(t *testing.T) {
		assert.Equal(t, c.C.SumObjPtr(&obj1, &obj2), s.SumObjPtr(&obj1, &obj2))
	})
	time1 := time.Now()
	time2 := time.Hour * 40
	t.Run("", func(t *testing.T) {
		assert.Equal(t, c.C.AddTime(time1, time2).Second(), s.AddTime(time1, time2).Second())
	})
	t.Run("", func(t *testing.T) {
		assert.Equal(t, c.C.AddTimePtr(&time1, &time2).Second(), s.AddTimePtr(&time1, &time2).Second())
	})
	t.Run("ConvertError", func(t *testing.T) {
		msg := "hello"
		result, err := c.C.ConvertError(msg)
		assert.Equal(t, result, "")
		assert.Equal(t, err.Error(), msg)
	})
	t.Run("ConvertNilError", func(t *testing.T) {
		result, err := c.C.ConvertError("")
		assert.Equal(t, result, "ok")
		assert.NoError(t, err)
	})
}
