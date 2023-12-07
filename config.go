package remote_ioc

type ServerConfig struct {
	Addr        string
	RoutePrefix string
}

type ClientConfig struct {
	Hosts []string
}
