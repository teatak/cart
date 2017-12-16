package logger

import (
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type Spliter int

const (
	EveryMin Spliter = 1 + iota
	EveryHour
	EveryDay
)

type FileWriter struct {
	dir     string
	spliter Spliter
	mu      sync.Mutex
	ticker  string
	file    io.WriteCloser
}

var _ io.Writer = &FileWriter{}

func DefaultFileWriter(dir string) *FileWriter {
	return NewFileWriter(dir, EveryDay)
}

func NewFileWriter(dir string, spliter Spliter) *FileWriter {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return &FileWriter{
		dir:     dir,
		spliter: spliter,
	}
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	var should string
	now := time.Now()
	switch w.spliter {
	case EveryMin:
		should = now.Format("LOG_200601021504.log")
		break
	case EveryHour:
		should = now.Format("LOG_2006010215.log")
		break
	case EveryDay:
		should = now.Format("LOG_20060102.log")
		break
	}
	if should != w.ticker {
		w.mu.Lock()
		if w.file != nil {
			w.file.Close()
		}
		//new file
		w.ticker = should
		w.file, _ = os.OpenFile(path.Join(w.dir, should), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		w.mu.Unlock()
	}
	return w.file.Write(p)
}
