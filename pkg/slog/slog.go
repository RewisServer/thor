// simple logging package to wrap
// the standard logger with different logging output
// streams (out/err) and the ability for verbose
// logging.
package slog

import (
	"fmt"
	"log"
	"os"
)

var (
	// verbosity level 0 is the standard non-verbose level.
	verbosity int8 = 0

	// loggers for info and error output
	stdout = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	stderr = log.New(os.Stderr, "", log.Ldate | log.Ltime)
)

// Verbose type which implements Info, Infoln and Infof.
// Will be used to determine, if verbosity level is within range
// to print out the log.
// Not used for error messages, as these are not influenced by verbosity.
type Verbose bool

// Sets the current verbosity level.
// Can be between -127 and 127 (including)
// Normally, you wouldn't set the verbosity to < 0,
// but in cases of disabling logging, you can.
func SetVerbosity(v int8) {
	verbosity = v
}

// Determines if the given verbosity is within verbose range.
// Returns the Verbose logging object.
func V(v uint8) Verbose {
	return Verbose(int8(v) <= verbosity)
}

// Checks if the current Verbose's verbosity level
// is within range and only if that is true, prints
// out the message with 'Print'
// No new-line character will be placed.
func (v Verbose) Info(o ...interface{}) {
	if !v {
		return
	}
	stdout.Print(o...)
}

// Checks if the current Verbose's verbosity level
// is within range and only if that is true, prints
// out the message with 'Println'
func (v Verbose) Infoln(o ...interface{}) {
	if !v {
		return
	}
	stdout.Println(o...)
}

// Checks if the current Verbose's verbosity level
// is within range and only if that is true, prints
// out the message with 'Println' and 'Sprintf'
// A new-line character will be added.
func (v Verbose) Infof(format string, o ...interface{}) {
	if !v {
		return
	}
	stdout.Println(fmt.Sprintf(format, o...))
}

func Info(o ...interface{}) {
	V(0).Info(o...)
}

func Infoln(o ...interface{}) {
	V(0).Infoln(o...)
}

func Infof(format string, o ...interface{}) {
	V(0).Infof(format, o...)
}

func Debug(o ...interface{}) {
	V(1).Info(o...)
}

func Debugln(o ...interface{}) {
	V(1).Infoln(o...)
}

func Debugf(format string, o ...interface{}) {
	V(1).Infof(format, o...)
}

func Error(o ...interface{}) {
	stderr.Print(o...)
}

func Errorln(o ...interface{}) {
	stderr.Println(o...)
}

func Errorf(format string, o ...interface{}) {
	stderr.Printf(format, o...)
}

func Fatal(o ...interface{}) {
	stderr.Fatal(o...)
}

func Fatalln(o ...interface{}) {
	stderr.Fatalln(o...)
}

func Fatalf(format string, o ...interface{}) {
	stderr.Fatalf(format, o...)
}
