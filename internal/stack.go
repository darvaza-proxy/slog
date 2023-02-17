package internal

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const (
	// MaxDepth is the maximum depth we will go in the stack.
	MaxDepth = 32
)

// Frame represents a function call on the call Stack.
// This implementation is heavily based on
// github.com/pkg/errors.Frame but all parts are resolved
// immediatelly for later consumption.
type Frame struct {
	pc    uintptr
	entry uintptr
	name  string
	file  string
	line  int
}

func frameForPC(pc uintptr) Frame {
	var entry uintptr
	var name string
	var file string
	var line int

	if fp := runtime.FuncForPC(pc - 1); fp != nil {
		entry = fp.Entry()
		name = fp.Name()
		file, line = fp.FileLine(pc)
	} else {
		name = "unknown"
		file = "unknown"
	}

	return Frame{
		pc:    pc,
		entry: entry,
		name:  name,
		file:  file,
		line:  line,
	}
}

// Name returns the name of the function.
func (f Frame) Name() string {
	return f.name
}

// File returns the file name of the source code
// corresponding to this Frame
func (f Frame) File() string {
	return f.file
}

// Line returns the file number on the source code
// corresponding to this Frame, or zero if unknown.
func (f Frame) Line() int {
	return f.line
}

// FileLine returns File name and Line separated by
// a colon, or only the filename if the Line isn't known
func (f Frame) FileLine() string {
	if f.line > 0 {
		return fmt.Sprintf("%s:%v", f.file, f.line)
	}

	return f.file
}

/* Format formats the frame according to the fmt.Formatter interface.
 *
 *	%s    source file
 *	%d    source line
 *	%n    function name
 *	%v    equivalent to %s:%d
 *
 * Format accepts flags that alter the printing of some verbs, as follows:
 *
 *	%+s   function name and path of source file relative to the compile time
 *	      GOPATH separated by \n\t (<funcname>\n\t<path>)
 *	%+n   full package name followed by function name
 *  %+v   equivalent to %+s:%d
 */
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name)
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file)
		default:
			io.WriteString(s, path.Base(f.file))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line))
	case 'n':
		n := f.name
		switch {
		case s.Flag('+'):
			io.WriteString(s, n)
		default:
			io.WriteString(s, funcname(n))
		}
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// Stack is an snapshot of the call stack in
// the form of an array of Frames.
type Stack []Frame

// Format formats the stack of Frames following the rules
// explained in Frame.Format with the adition of the '#' flag.
//
// when '#' is passed, like for example %#+v each row
// will be prefixed by [i/n] indicating the position in the stack
// followed by the %+v representation of the Frame
func (st Stack) Format(s fmt.State, verb rune) {
	if s.Flag('#') {
		l := len(st)
		for i, f := range st {
			fmt.Fprintf(s, "\n[%v/%v] ", i, l)
			f.Format(s, verb)
		}
	} else {
		for _, f := range st {
			io.WriteString(s, "\n")
			f.Format(s, verb)
		}
	}
}

func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// Here returns the Frame corresponding to where it was called,
// or nil if it wasn't possible
func Here() *Frame {
	const depth = 1
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > 0 {
		f := frameForPC(pcs[0])
		return &f
	}
	return nil
}

// StackFrame returns the Frame skip levels above from where it
// was called, or nil if it wasn't possible
func StackFrame(skip int) *Frame {
	const depth = MaxDepth
	var pcs [depth]uintptr

	if n := runtime.Callers(2, pcs[:]); n > skip {
		f := frameForPC(pcs[skip])
		return &f
	}

	return nil
}

// StackTrace returns a snapshot of the call stack starting
// skip levels above from where it was called, on an empty
// array if it wasn't possible
func StackTrace(skip int) Stack {
	const depth = MaxDepth
	var pcs [depth]uintptr
	var st Stack

	if n := runtime.Callers(2, pcs[:]); n > skip {
		var frames []Frame

		n = n - skip
		frames = make([]Frame, 0, n)

		for _, pc := range pcs[skip:n] {
			frames = append(frames, frameForPC(pc))
		}

		st = Stack(frames)
	}

	return st
}
