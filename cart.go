package cart

/*
A HTTP web framework written in golang

	 ██████╗ █████╗ ██████╗ ████████╗
	██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
	██║     ███████║██████╔╝   ██║
	██║     ██╔══██║██╔══██╗   ██║
	╚██████╗██║  ██║██║  ██║   ██║
	 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
*/
import (
	"github.com/teatak/cart/render"
	"html/template"
)

const Version = "v1.0.8"

func New() *Engine {
	debugWarning()
	e := &Engine{
		Router:              Router{Path: "/"},
		ForwardedByClientIP: true,
		AppEngine:           false,
		delims:              render.Delims{Left: "{{", Right: "}}"},
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
