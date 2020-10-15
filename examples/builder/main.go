package main

import (
	"bytes"

	"github.com/nikandfor/tlog"
)

func main() {
	var buf bytes.Buffer

	l := tlog.New(tlog.NewJSONWriter(&buf), tlog.NewConsoleWriter(tlog.Stderr, tlog.LdetFlags))

	// usual way
	l.Printw("message", tlog.AInt("int", 100), tlog.AStr("str", "string"))

	// the same output but customizable
	l.BuildMessage().Now().Location(tlog.Caller(0)).Int("int", 100).Str("str", "string").Printf("message")

	// empty event
	l.BuildMessage().Printf("")

	// without time
	tr := l.BuildSpanStart().NewID().Caller(0).Start()

	// without location
	tr.BuildMessage().Now().Printf("message")

	hotCode(tr, 300)

	tr.Finish()

	_, _ = buf.WriteTo(tlog.Stderr)
}

var hotCodeLoc tlog.PC

func hotCode(tr tlog.Span, arg int) {
	tr.BuildMessage().
		Now().                      // current time
		CallerOnce(0, &hotCodeLoc). // faster than Caller
		Printf("arg: %v", arg)      // Printf is more efficient than Printw
}
