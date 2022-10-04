package prompt

import (
	"bytes"
	"fmt"
	"time"
)

type Context struct {
	buf *bytes.Buffer
}

func With(count int64, now time.Time, addr string) Context {
	const TimeFmt = "15:04:05"
	var buf = bytes.NewBufferString("")
	fmt.Fprintf(buf, "> [%03v][%v] %v:", count, now.Format(TimeFmt), addr)
	return Context{buf: buf}
}

func (c Context) WithTakeTime(takeTime time.Duration) Context {
	fmt.Fprintf(c.buf, "%v ms", takeTime.Milliseconds())
	return c
}

func (c Context) WithText(text string) Context {
	fmt.Fprintf(c.buf, "%s", text)
	return c
}

func (c Context) WithBlank() Context {
	fmt.Fprint(c.buf, " ")
	return c
}

func (c Context) WithError(err error) Context {
	fmt.Fprintf(c.buf, "%v", err)
	return c
}

func (c Context) String() string {
	return c.buf.String()
}
