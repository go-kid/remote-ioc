package defination

import (
	"time"
)

type RemoteComponent interface {
	RemoteServiceId() string
}

type RemoteMethodExport interface {
	ExportMethods() []string
}

type RemoteMethodExclude interface {
	ExcludeMethods() []string
}

type ServerInfo struct {
	Addr  string
	Delay time.Duration
}

type LoadBalancing func(servers []*ServerInfo) int

type Invoke func(methodName string, v ...any) ([]any, error)

type InvokeComponent interface {
	RemoteServiceId() string
	RegisterInvoker(invoke Invoke)
}

type LoadBalancingStrategy interface {
	LoadBalancing() LoadBalancing
}
