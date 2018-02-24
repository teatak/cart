package render

import (
	"fmt"
	"io"
	"net/http"
)

type String struct {
	Format string
	Data   []interface{}
}

var plainContentType = []string{"text/plain; charset=utf-8"}

func (r String) Render(w http.ResponseWriter) error {
	WriteString(w, r.Format, r.Data)
	return nil
}

func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

func WriteString(w http.ResponseWriter, format string, data []interface{}) {
	writeContentType(w, plainContentType)
	if len(data) > 0 {
		fmt.Fprintf(w, format, data...)
	} else {
		io.WriteString(w, format)
	}
}
