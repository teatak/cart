package cart

import (
	"path"
	"os"
	"encoding/xml"
)

type Next func()
//normal handler
type Handler func(*Context, Next)
type HandlerFinal func(*Context)
type HandlerCompose func(*Context, Next) Next

type H map[string]interface{}
// MarshalXML allows type H to be used with xml.Marshal
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return nil
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); len(port) > 0 {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too much parameters")
	}
}

func lastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if len(relativePath) == 0 {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}
/*
transfer Handler to HandlerCompose func
 */
func makeCompose(handles ...Handler) HandlerCompose {
	composeHandles := []HandlerCompose{}
	for _, handle := range handles {
		innerHandle := handle
		tempHandle := func(c *Context, next Next) Next {
			return func() {
				innerHandle(c,next)
			}
		}
		composeHandles = append(composeHandles, tempHandle)
	}
	return compose(composeHandles...)
}

/*
compose HandlerCompose
	temp := 0
	A := func(c *Context, next Next) Next {
		return func() {
			temp = temp + 2;
			next()
		}
	}
	B := func(c *Context, next Next) Next {
		return func() {
			temp = temp * 2;
			next()
		}
	}
	composed := compose(A,B,B)(nil, func(){
		//this is the end the temp value is (0+2)*2*2
	})
	composed()
 */
func compose(functions ...HandlerCompose) HandlerCompose {
	if len(functions) == 0 {
		return nil
	}
	if len(functions) == 1 {
		return functions[0]
	}

	return func(c *Context, next Next) Next {
		last := functions[len(functions)-1]
		rest := functions[0:len(functions)-1]
		composed := last(c, next);
		for i, _ := range rest {
			composed = rest[len(rest)-1-i](c, composed)
		}
		return composed
	}
}
