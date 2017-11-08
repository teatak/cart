package cart

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestColorForMethod(t *testing.T) {
	Equal(t, colorForMethod("GET"), string([]byte{27, 91, 57, 55, 59, 52, 52, 109}), "get should be blue")
	Equal(t, colorForMethod("POST"), string([]byte{27, 91, 57, 55, 59, 52, 54, 109}), "post should be cyan")
	Equal(t, colorForMethod("PUT"), string([]byte{27, 91, 57, 55, 59, 52, 51, 109}), "put should be yellow")
	Equal(t, colorForMethod("DELETE"), string([]byte{27, 91, 57, 55, 59, 52, 49, 109}), "delete should be red")
	Equal(t, colorForMethod("PATCH"), string([]byte{27, 91, 57, 55, 59, 52, 50, 109}), "patch should be green")
	Equal(t, colorForMethod("HEAD"), string([]byte{27, 91, 57, 55, 59, 52, 53, 109}), "head should be magenta")
	Equal(t, colorForMethod("OPTIONS"), string([]byte{27, 91, 57, 48, 59, 52, 55, 109}), "options should be white")
	Equal(t, colorForMethod("TRACE"), string([]byte{27, 91, 48, 109}), "trace is not defined and should be the reset color")
}

func TestColorForStatus(t *testing.T) {
	Equal(t, colorForStatus(200), string([]byte{27, 91, 57, 55, 59, 52, 50, 109}), "2xx should be green")
	Equal(t, colorForStatus(301), string([]byte{27, 91, 57, 48, 59, 52, 55, 109}), "3xx should be white")
	Equal(t, colorForStatus(404), string([]byte{27, 91, 57, 55, 59, 52, 51, 109}), "4xx should be yellow")
	Equal(t, colorForStatus(2), string([]byte{27, 91, 57, 55, 59, 52, 49, 109}), "other things should be red")
}

func TestIsDebugging(t *testing.T) {
	SetMode(DebugMode)
	if !IsDebugging() {
		t.Errorf("Wrong return IsDebugging should true but return %v", IsDebugging())
	}
	SetMode(ReleaseMode)
	if IsDebugging() {
		t.Errorf("Wrong return IsDebugging should false but return %v", IsDebugging())
	}
}

func TestDebugPrint(t *testing.T) {
	var w bytes.Buffer
	setup(&w)
	defer teardown()

	SetMode(DebugMode)
	disableColor = true

	debugError(errors.New("new error"))

	debugPrint("these are |%d %s\n", 2, "error messages")
	//w.String()
	s := strings.Split(w.String(), "|")[1]
	if s != "2 error messages\n" {
		t.Errorf("Wrong return debugPrint %s", w.String())
	}

	disableColor = false

	debugError(errors.New("new error"))
	debugPrint("these are |%d %s\n", 2, "error messages")
}

//utils

func setup(w io.Writer) {
	SetMode(DebugMode)
	log.SetOutput(w)
}

func teardown() {
	SetMode(DebugMode)
	log.SetOutput(os.Stdout)
}
