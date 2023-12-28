package server

import (
	"github.com/go-kid/remote-ioc/http/transmission"
)

type Config struct {
	Addr                   string
	RoutePrefix            string
	SerializationFilters   []SerializationFilter
	DeserializationFilters []DeserializationFilter
}

type DeserializationFilter = transmission.DeserializationFilter

type SerializationFilter = transmission.SerializationFilter
