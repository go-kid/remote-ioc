package remote_ioc

import "github.com/go-kid/ioc/app"

func Handle(c ServerConfig) app.SettingOption {
	return func(s *app.App) {
		s.Register(&iocServer{
			c: c,
			r: s.Registry,
		})
	}
}

func Remote(c ClientConfig) app.SettingOption {
	return func(s *app.App) {
		s.Registry.Register(&iocClient{
			c: c,
			r: s.Registry,
		})
	}
}
