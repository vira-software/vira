package vira

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

const ginSupportMinGoVer = 18

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(vira.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return viraMode == debugCode
}

// DebugPrintRouteFunc indicates debug log output format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

// DebugPrintFunc indicates debug log output format.
var DebugPrintFunc func(format string, values ...interface{})

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())
		if DebugPrintRouteFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrint(format string, values ...any) {
	if !IsDebugging() {
		return
	}

	if DebugPrintFunc != nil {
		DebugPrintFunc(format, values...)
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(DefaultWriter, "[VIRA-debug] "+format, values...)
}

func getMinVer(v string) (uint64, error) {
	first := strings.IndexByte(v, '.')
	last := strings.LastIndexByte(v, '.')
	if first == last {
		return strconv.ParseUint(v[first+1:], 10, 64)
	}
	return strconv.ParseUint(v[first+1:last], 10, 64)
}

func debugPrintWARNINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
		debugPrint(`[WARNING] Now Vira requires Go 1.18+.

`)
	}
	debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

`)
}

func debugPrintWARNINGNew() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export VIRA_MODE=release
 - using code:	vira.SetMode(vira.ReleaseMode)

`)
}

func debugPrintError(err error) {
	if err != nil && IsDebugging() {
		fmt.Fprintf(DefaultErrorWriter, "[VIRA-debug] [ERROR] %v\n", err)
	}
}
