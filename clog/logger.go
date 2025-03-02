package clog

import (
	"io"
	"log"
	"os"
)

var std = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)

type any = interface{}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

func Flags() int {
	return std.Flags()
}

func SetFlags(flag int) {
	std.SetFlags(flag)
}

func Prefix() string {
	return std.Prefix()
}

func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

func Writer() io.Writer {
	return std.Writer()
}

func Print(v ...any) {
	std.Print(v...)
}

func Printf(format string, v ...any) {
	std.Printf(format, v...)
}

func Println(v ...any) {
	std.Println(v...)
}

func Fatal(v ...any) {
	std.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	std.Fatalf(format, v...)
}

func Fatalln(v ...any) {
	std.Fatalln(v...)
}

func Panic(v ...any) {
	std.Panic(v...)
}

func Panicf(format string, v ...any) {
	std.Panicf(format, v...)
}

func Panicln(v ...any) {
	std.Panicln(v...)
}

func Output(calldepth int, s string) error {
	return std.Output(calldepth, s)
}
