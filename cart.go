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
	e := &Engine{}
	e.init()
	e.Use("/",Logger())
	return e
}