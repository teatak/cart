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
	beforeFuncs []func()
	afterFuncs  []func()
	size        int
	status      int
}

// var _ ResponseWriter = &responseWriter{}

func (w *ResponseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.beforeFuncs = nil
	w.afterFuncs = nil
	w.size = noWritten
	w.status = defaultStatus
}

func (w *ResponseWriter) Before(fn func()) {
	w.beforeFuncs = append(w.beforeFuncs, fn)
}

func (w *ResponseWriter) After(fn func()) {
	w.afterFuncs = append(w.afterFuncs, fn)
}

func (w *ResponseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}
		w.status = code
	}
}

func (w *ResponseWriter) WriteHeaderFinal() {
	if !w.Written() {
		for _, fn := range w.beforeFuncs {
			fn()
		}
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
		for _, fn := range w.afterFuncs {
			fn()
		}
	}
}

func (w *ResponseWriter) writeHeader() {
	if !w.Written() {
		for _, fn := range w.beforeFuncs {
			fn()
		}
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *ResponseWriter) Write(data []byte) (n int, err error) {
	w.writeHeader()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	for _, fn := range w.afterFuncs {
		fn()
	}
	return
}

func (w *ResponseWriter) WriteString(s string) (n int, err error) {
	w.writeHeader()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	for _, fn := range w.afterFuncs {
		fn()
	}
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
	w.ResponseWriter.(http.Flusher).Flush()
}
