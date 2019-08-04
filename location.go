package tlog

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

// Location is a program counter alias.
// Function name, file name and line could be obtained from it but only in the same binary where Caller of Funcentry was called.
type Location uintptr

// Caller returns information about the calling goroutine's stack. The argument s is the number of frames to ascend, with 0 identifying the caller of Caller.
//
// It's hacked version of runtime.Caller with no allocs.
func Caller(s int) Location {
	var pc [1]uintptr
	runtime.Callers(2+s, pc[:])
	return Location(pc[0])
}

// Funcentry returns information about the calling goroutine's stack. The argument s is the number of frames to ascend, with 0 identifying the caller of Caller.
//
// It's hacked version of runtime.Callers -> runtime.CallersFrames -> Frames.Next -> Frame.Entry with no allocs.
func Funcentry(s int) Location {
	var pc [1]uintptr
	runtime.Callers(2+s, pc[:])
	return Location(Location(pc[0]).Entry())
}

// String formats Location as base_name.go:line.
// Works only in the same binary where Caller of Funcentry was called.
func (l Location) String() string {
	_, file, line := l.NameFileLine()
	return fmt.Sprintf("%v:%d", path.Base(file), line)
}

func cropFilename(fn, tp string) string {
	p := strings.LastIndexByte(tp, '/')
	pp := strings.IndexByte(tp[p+1:], '.')
	tp = tp[:p+pp]

again:
	if p = strings.Index(fn, tp); p != -1 {
		return fn[p:]
	}

	p = strings.IndexByte(tp, '/')
	if p == -1 {
		return path.Base(fn)
	}
	tp = tp[p+1:]
	goto again
}
