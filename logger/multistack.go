package logger

import (
	"strings"

	"github.com/pkg/errors"
)

type StackTrace struct {
	Error  string  `json:"error"`
	Frames []frame `json:"stacktrace"`
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
type frame struct {
	StackSourceFileName string `json:"source"`
	StackSourceLineName string `json:"line"`
	StackSourceFuncName string `json:"func"`
}

func MarshalMultiStack(err error) interface{} {
	stackTraces := []StackTrace{}
	currentErr := err
	for currentErr != nil {
		stack, ok := currentErr.(stackTracer)
		if !ok {
			currentErr = unwrapErr(currentErr)
			continue
		}
		st := stack.StackTrace()
		s := &formatState{}
		stackTrace := StackTrace{
			Error: currentErr.Error(),
		}

		for _, f := range st {
			frame := frame{
				StackSourceFileName: frameField(f, &formatState{b: []byte{'+'}}, 's'),
				StackSourceLineName: frameField(f, s, 'd'),
				StackSourceFuncName: frameField(f, s, 'n'),
			}
			switch {
			case strings.HasSuffix(frame.StackSourceFileName, "runtime/asm_amd64.s"):
			case strings.HasSuffix(frame.StackSourceFileName, "runtime/proc.go"):
			default:
				stackTrace.Frames = append(stackTrace.Frames, frame)
			}
		}
		stackTraces = append(stackTraces, stackTrace)

		currentErr = unwrapErr(currentErr)
	}
	return stackTraces
}

func unwrapErr(err error) error {
	cause, ok := err.(causer)
	if !ok {
		return nil
	}
	return cause.Cause()
}

type causer interface {
	Cause() error
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
