package cart

const Version = "v1.0"

func New() *Engine {
	debugWarning()
	e := &Engine{}
	e.init()
	e.Use("/",Logger())
	return e
}