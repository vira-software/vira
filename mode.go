package vira

import (
	"flag"
	"io"
	"os"

	"github.com/vira-software/vira/binding"
)

// EnvGinMode indicates environment name for vira mode.
const EnvGinMode = "VIRA_MODE"

const (
	// DebugMode indicates vira mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates vira mode is release.
	ReleaseMode = "release"
	// TestMode indicates vira mode is test.
	TestMode = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

// DefaultWriter is the default io.Writer used by Vira for debug output and
// middleware output like Logger() or Recovery().
// Note that both Logger and Recovery provides custom ways to configure their
// output io.Writer.
// To support coloring in Windows use:
//
//	import "github.com/mattn/go-colorable"
//	vira.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Vira to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

var (
	viraMode = debugCode
	modeName = DebugMode
)

func init() {
	mode := os.Getenv(EnvGinMode)
	SetMode(mode)
}

// SetMode sets vira mode according to input string.
func SetMode(value string) {
	if value == "" {
		if flag.Lookup("test.v") != nil {
			value = TestMode
		} else {
			value = DebugMode
		}
	}

	switch value {
	case DebugMode:
		viraMode = debugCode
	case ReleaseMode:
		viraMode = releaseCode
	case TestMode:
		viraMode = testCode
	default:
		panic("vira mode unknown: " + value + " (available mode: debug release test)")
	}

	modeName = value
}

// DisableBindValidation closes the default validator.
func DisableBindValidation() {
	binding.Validator = nil
}

// EnableJsonDecoderUseNumber sets true for binding.EnableDecoderUseNumber to
// call the UseNumber method on the JSON Decoder instance.
func EnableJsonDecoderUseNumber() {
	binding.EnableDecoderUseNumber = true
}

// EnableJsonDecoderDisallowUnknownFields sets true for binding.EnableDecoderDisallowUnknownFields to
// call the DisallowUnknownFields method on the JSON Decoder instance.
func EnableJsonDecoderDisallowUnknownFields() {
	binding.EnableDecoderDisallowUnknownFields = true
}

// Mode returns current vira mode.
func Mode() string {
	return modeName
}
