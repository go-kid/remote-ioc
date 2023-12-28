package http

import (
	"context"
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/remote-ioc/http/client"
	"github.com/go-kid/remote-ioc/http/server"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func startServer(t *testing.T, port int) {
	var s = &ServerComponentImpl{}
	ioc.RunTest(t,
		app.SetComponents(s),
		server.Handle(server.Config{
			Addr:        fmt.Sprintf(":%d", port),
			RoutePrefix: "",
		}),
	)
}

func TestRemoteIOC(t *testing.T) {
	var ports []int
	for i := 8888; i < 8888+10; i++ {
		startServer(t, i)
		ports = append(ports, i)
	}
	var s = &ServerComponentImpl{}

	var c = &ClientApp{}
	ioc.RunTest(t,
		app.SetComponents(c, &ServerComponentInvoker{}),
		client.Remote(client.Config{
			Servers: lo.Map(ports, func(item int, index int) client.ServerConfig {
				return client.ServerConfig{
					Addr:        fmt.Sprintf("http://localhost:%d", item),
					RoutePrefix: "",
				}
			}),
			Debug: false,
			LoadBalance: func(servers []*client.ServerInfo) int {
				var (
					minIndex int
					minDur   = servers[0].Delay
				)
				for i := 0; i < len(servers); i++ {
					if delay := servers[i].Delay; delay < minDur {
						minDur = delay
						minIndex = i
					}
					fmt.Println(servers[i])
				}
				fmt.Println("min", minIndex)
				return minIndex
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
	t.Run("AddTime", func(t *testing.T) {
		assert.Equal(t, c.C.AddTime(time1, time2).Second(), s.AddTime(time1, time2).Second())
	})
	t.Run("AddTimePtr", func(t *testing.T) {
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
	t.Run("WithContext", func(t *testing.T) {
		ctx := context.WithValue(
			context.WithValue(context.Background(),
				"key", "val123",
			), 123, "123",
		)
		result := c.C.WithContext(ctx)
		assert.Equal(t, result, "ok")
	})
}
