package render

import "net/http"

type Data struct {
	ContentType string
	Data        []byte
}

// Render (Data) writes data with custom ContentType
func (r Data) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	_, err = w.Write(r.Data)
	return
}

func (r Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
}
