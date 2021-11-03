package stack

import (
	"fmt"
	"go/build"
	"os"
	"runtime"
	"strings"
)

var (
	wd     string
	goROOT = build.Default.GOROOT
	goPath = build.Default.GOPATH
	goStd  = goROOT + "/src"
)

func init() {
	wd, _ = os.Getwd()
	if wd != "" {
		wd += "/"
	}
}

// Caller Caller
func Caller(depth int) string {
	pc, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		return ""
	}
	function := trimFunction(runtime.FuncForPC(pc).Name())
	return fmt.Sprintf("%s@%s:%d", function, trimFile(file), line)
}

// Callers Callers
func Callers(filters ...func(string) bool) string {
	filter := combineFilter(append(filters, stackfilter)...)
	pcs := make([]uintptr, 1024)
	pcs = pcs[:runtime.Callers(0, pcs)]
	frames := runtime.CallersFrames(pcs)
	stack := make([]string, 0, len(pcs))
next:
	frame, hasNext := frames.Next()
	if hasNext {
		function := trimFunction(frame.Function)
		if filter(frame.File) {
			goto next
		}
		file := trimFile(frame.File)
		if filter(file) {
			goto next
		}
		line := frame.Line
		stack = append(stack, fmt.Sprintf("%s@%s:%d", function, file, line))
		goto next
	}
	return strings.Join(stack, ",")
}

func combineFilter(filters ...func(string) bool) func(string) bool {
	return func(file string) bool {
		for _, filter := range filters {
			if filter != nil && filter(file) {
				return true
			}
		}
		return false
	}
}

func trimFunction(src string) string {
	return src[strings.LastIndexByte(src, '.')+1:]
}

func trimFile(src string) string {
	src = strings.TrimPrefix(src, goPath+"/pkg/mod/")
	return strings.TrimPrefix(src, wd)
}

func stackfilter(src string) bool {
	return strings.Contains(src, "github.com/wwq-2020/go.common")
}

// StdFilter StdFilters
func StdFilter(src string) bool {
	return strings.Contains(src, goStd)
}
