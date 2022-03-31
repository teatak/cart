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
