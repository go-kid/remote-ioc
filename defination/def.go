package defination

type RemoteComponent interface {
	RemoteServiceId() string
}

type RemoteMethodExport interface {
	ExportMethods() []string
}

type RemoteMethodExclude interface {
	ExcludeMethods() []string
}

type Invoke func(methodName string, v ...any) ([]any, error)

type InvokeComponent interface {
	RemoteServiceId() string
	RegisterInvoker(invoke Invoke)
}
