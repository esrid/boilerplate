package server

import "net/http"

type Server interface {
	Start() error
}

type Router interface {
	Handler() http.Handler
}

type Middleware func(http.Handler) http.Handler
