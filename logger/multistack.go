package logger

import (
	"strings"

	"github.com/pkg/errors"
)

type frame struct {
	File string `json:"file"`
	Func string `json:"func"`
	Line string `json:"line"`
}

func marshallStack(err error) interface{} {
	errStackTracer, ok := errors.Cause(err).(stackTracer)
	if !ok {
		return nil
	}

	st := errStackTracer.StackTrace()

	frames := make([]frame, 0, len(st))

	for _, f := range st {
		fs := &formatState{}

		resFrame := frame{
			File: frameField(f, &formatState{b: []byte{'+'}}, 's'),
			Func: frameField(f, fs, 'n'),
			Line: frameField(f, fs, 'd'),
		}

		switch {
		case strings.HasSuffix(resFrame.File, "runtime/asm_amd64.s"):
		case strings.HasSuffix(resFrame.File, "runtime/proc.go"):
		default:
			frames = append(frames, resFrame)
		}
	}
	return frames
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type formatState struct {
	b []byte
}

func (s *formatState) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

func (s *formatState) Width() (wid int, ok bool) {
	return 0, false
}

func (s *formatState) Precision() (prec int, ok bool) {
	return 0, false
}

func (s *formatState) Flag(c int) bool {
	return string((*s).b) == "+"
}

func frameField(f errors.Frame, s *formatState, c rune) string {
	f.Format(s, c)
	return string(s.b)
}
