package client

import "github.com/go-kid/ioc/app"

func Remote(c Config) app.SettingOption {
	return func(s *app.App) {
		s.Registry.Register(&iocClient{
			c: c,
			r: s.Registry,
		})
	}
}
