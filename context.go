package cart

import "net/http"

type Context struct {
	Request   *http.Request
	Writer    http.ResponseWriter
	Keys     map[string]interface{}
}

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