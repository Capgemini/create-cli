package log

import (
	"fmt"
	"io"
)

type StderrLogger struct {
	Stderr io.Writer
	Tool   string
}

func (l StderrLogger) Actionf(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `►`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}

func (l StderrLogger) Generatef(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `✚`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}

func (l StderrLogger) Waitingf(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `◎`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}

func (l StderrLogger) Successf(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `✔`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}

func (l StderrLogger) Warningf(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `⚠️`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}

func (l StderrLogger) Failuref(format string, a ...interface{}) {
	fmt.Fprintln(l.Stderr, `✗`, " "+l.Tool+"›", fmt.Sprintf(format, a...))
}
