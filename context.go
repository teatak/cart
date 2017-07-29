package cart

import "net/http"

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(name string) (string, bool) {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return "", false
}

type Context struct {
	response 	responseWriter
	Request   	*http.Request
	Response    ResponseWriter

	Router		*Router
	Params   	Params

	Keys     	map[string]interface{}
}

/*
reset Con
 */
func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.response.reset(w)
	c.Response = &c.response
	c.Request = req
	c.Params = c.Params[0:0]
	c.Keys = nil
}

