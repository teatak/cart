package cart

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type ResponseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

// var _ ResponseWriter = &responseWriter{}

func (w *ResponseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *ResponseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			if IsDebugging() {
				debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			}
			return
		}
		w.status = code
	}
}

func (w *ResponseWriter) WriteHeaderFinal() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *ResponseWriter) writeHeader() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *ResponseWriter) Write(data []byte) (n int, err error) {
	w.writeHeader()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *ResponseWriter) WriteString(s string) (n int, err error) {
	w.writeHeader()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

func (w *ResponseWriter) Status() int {
	return w.status
}

func (w *ResponseWriter) Size() int {
	return w.size
}

func (w *ResponseWriter) Written() bool {
	return w.size != noWritten
}

// Implements the http.Hijacker interface
func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Implements the http.CloseNotify interface
// func (w *ResponseWriter) CloseNotify() <-chan bool {
// 	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
// }

// Implements the http.Flush interface
func (w *ResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
