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
	"html/template"

	"github.com/teatak/cart/v2/render"
)

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
