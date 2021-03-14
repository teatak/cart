/*
A HTTP web framework written in golang

	 ██████╗ █████╗ ██████╗ ████████╗
	██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
	██║     ███████║██████╔╝   ██║
	██║     ██╔══██║██╔══██╗   ██║
	╚██████╗██║  ██║██║  ██║   ██║
	 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
*/
package cart

import (
	"github.com/teatak/cart/render"
	"html/template"
)

const Version = "v1.0.7"

func New() *Engine {
	debugWarning()
	e := &Engine{
		Router:              Router{Path: "/"},
		ForwardedByClientIP: true,
		AppEngine:           false,
		delims:              render.Delims{"{{", "}}"},
		FuncMap:             template.FuncMap{},
	}

	e.init()
	return e
}

func Default() *Engine {
	e := New()
	e.Use("/", Logger(), RecoveryRender(DefaultErrorWriter))
	return e
}
