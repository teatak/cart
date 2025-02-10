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

type (
	ResponseWriter interface {
		http.ResponseWriter
		http.Hijacker
		http.Flusher
		Size() int
		Status() int
		Written() bool
	}
	responseWriter struct {
		http.ResponseWriter
		beforeFuncs []func()
		afterFuncs  []func()
		size        int
		status      int
		written     bool
	}
)

var _ ResponseWriter = &responseWriter{}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.beforeFuncs = nil
	w.afterFuncs = nil
	w.size = noWritten
	w.status = defaultStatus
	w.written = false
}

func (w *responseWriter) Before(fn func()) {
	w.beforeFuncs = append(w.beforeFuncs, fn)
}

func (w *responseWriter) After(fn func()) {
	w.afterFuncs = append(w.afterFuncs, fn)
}

func (w *responseWriter) WriteHeader(code int) {
	if w.written {
		debugPrint("[WARNING] Headers were already written.")
		return
	}
	w.status = code
	for _, fn := range w.beforeFuncs {
		fn()
	}
	w.ResponseWriter.WriteHeader(w.status)
	w.written = true
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	if !w.written {
		w.WriteHeader(w.status)
	}
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	for _, fn := range w.afterFuncs {
		fn()
	}
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	if !w.written {
		w.WriteHeader(w.status)
	}
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	for _, fn := range w.afterFuncs {
		fn()
	}
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.written
}

// Implements the http.Hijacker interface
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Implements the http.CloseNotify interface
// func (w *responseWriter) CloseNotify() <-chan bool {
// 	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
// }

// Implements the http.Flush interface
func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}
