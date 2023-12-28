package server

import "github.com/go-kid/ioc/app"

func Handle(c Config) app.SettingOption {
	return func(s *app.App) {
		s.Register(&iocServer{
			c: c,
			r: s.Registry,
		})
	}
}
