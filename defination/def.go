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

type RemoteMethodReplace interface {
	ReplaceMethods() map[string]string
}
