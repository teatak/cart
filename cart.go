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

const Version = "v1.0.0"

func New() *Engine {
	debugWarning()
	e := &Engine{
		Router: Router{Path:"/"},
		ForwardedByClientIP:    true,
		AppEngine:              false,
	}

	e.init()
	return e
}

func Default() *Engine {
	e := New()
	e.Use("/",Logger(),RecoveryRender(DefaultErrorWriter))
	return e
}