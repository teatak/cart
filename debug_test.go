package cart

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
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
	// w.String()
	if !strings.Contains(w.String(), "these are |2 error messages") {
		t.Errorf("Wrong return debugPrint %s", w.String())
	}

	disableColor = false

	debugError(errors.New("new error"))
	debugPrint("these are |%d %s\n", 2, "error messages")
}

//utils

func setup(w io.Writer) {
	SetMode(DebugMode)
	slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func teardown() {
	SetMode(DebugMode)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}
