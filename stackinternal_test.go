package stack

import (
	"runtime"
	"testing"
)

func TestCaller(t *testing.T) {
	t.Parallel()

	c := Caller(0)
	_, file, line, ok := runtime.Caller(0)
	line--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	if got, want := c.frame.File, file; got != want {
		t.Errorf("got file == %v, want file == %v", got, want)
	}

	if got, want := c.frame.Line, line; got != want {
		t.Errorf("got line == %v, want line == %v", got, want)
	}
}

func f3(f1 func() Call) Call {
	return f2(f1)
}

func f2(f1 func() Call) Call {
	return f1()
}

func TestCallerMidstackInlined(t *testing.T) {
	t.Parallel()

	_, _, line, ok := runtime.Caller(0)
	line -= 10 // adjust to return f1() line inside f2()
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	c := f3(func() Call {
		return Caller(2)
	})

	if got, want := c.frame.Line, line; got != want {
		t.Errorf("got line == %v, want line == %v", got, want)
	}
	if got, want := c.frame.Function, "github.com/go-stack/stack.f3"; got != want {
		t.Errorf("got func name == %v, want func name == %v", got, want)
	}
}

func TestCallerPanic(t *testing.T) {
	t.Parallel()

	var (
		line int
		ok   bool
	)

	defer func() {
		if recover() != nil {
			var pcs [32]uintptr
			n := runtime.Callers(1, pcs[:])
			frames := runtime.CallersFrames(pcs[:n])
			// count frames to runtime.sigpanic
			panicIdx := 0
			for {
				f, more := frames.Next()
				if f.Function == "runtime.sigpanic" {
					break
				}
				panicIdx++
				if !more {
					t.Fatal("no runtime.sigpanic entry on the stack")
				}
			}
			if got, want := Caller(panicIdx).frame.Function, "runtime.sigpanic"; got != want {
				t.Errorf("sigpanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).frame.Function, "github.com/go-stack/stack.TestCallerPanic"; got != want {
				t.Errorf("TestCallerPanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).frame.Line, line; got != want {
				t.Errorf("TestCallerPanic frame: got line == %v, want line == %v", got, want)
			}
		}
	}()

	_, _, line, ok = runtime.Caller(0)
	line += 7 // adjust to match line of panic below
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// Initiate a sigpanic.
	var x *uintptr
	_ = *x
}

func TestCallerInlinedPanic(t *testing.T) {
	t.Parallel()

	var line int

	defer func() {
		if recover() != nil {
			var pcs [32]uintptr
			n := runtime.Callers(1, pcs[:])
			frames := runtime.CallersFrames(pcs[:n])
			// count frames to runtime.sigpanic
			panicIdx := 0
			for {
				f, more := frames.Next()
				if f.Function == "runtime.sigpanic" {
					break
				}
				panicIdx++
				if !more {
					t.Fatal("no runtime.sigpanic entry on the stack")
				}
			}
			if got, want := Caller(panicIdx).frame.Function, "runtime.sigpanic"; got != want {
				t.Errorf("sigpanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).frame.Function, "github.com/go-stack/stack.inlinablePanic"; got != want {
				t.Errorf("TestCallerInlinedPanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).frame.Line, line; got != want {
				t.Errorf("TestCallerInlinedPanic frame: got line == %v, want line == %v", got, want)
			}
		}
	}()

	doPanic(t, &line)
	t.Fatal("failed to panic")
}

func doPanic(t *testing.T, panicLine *int) {
	_, _, line, ok := runtime.Caller(0)
	*panicLine = line + 11 // adjust to match line of panic below
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	inlinablePanic()
}

func inlinablePanic() {
	// Initiate a sigpanic.
	var x *uintptr
	_ = *x
}

type tholder struct {
	trace func() CallStack
}

func (th *tholder) traceLabyrinth() CallStack {
	for {
		return th.trace()
	}
}

func TestTrace(t *testing.T) {
	t.Parallel()

	_, _, line, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	fh := tholder{
		trace: func() CallStack {
			cs := Trace()
			return cs
		},
	}

	cs := fh.traceLabyrinth()

	lines := []int{line + 7, line - 7, line + 12}

	for i, line := range lines {
		if got, want := cs[i].frame.Line, line; got != want {
			t.Errorf("got line[%d] == %v, want line[%d] == %v", i, got, i, want)
		}
	}
}

// Test stack handling originating from a sigpanic.
func TestTracePanic(t *testing.T) {
	t.Parallel()

	var (
		line int
		ok   bool
	)

	defer func() {
		if recover() != nil {
			trace := Trace()

			// find runtime.sigpanic
			panicIdx := -1
			for i, c := range trace {
				if c.frame.Function == "runtime.sigpanic" {
					panicIdx = i
					break
				}
			}
			if panicIdx == -1 {
				t.Fatal("no runtime.sigpanic entry on the stack")
			}
			if got, want := trace[panicIdx].frame.Function, "runtime.sigpanic"; got != want {
				t.Errorf("sigpanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := trace[panicIdx+1].frame.Function, "github.com/go-stack/stack.TestTracePanic"; got != want {
				t.Errorf("TestTracePanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := trace[panicIdx+1].frame.Line, line; got != want {
				t.Errorf("TestTracePanic frame: got line == %v, want line == %v", got, want)
			}
		}
	}()

	_, _, line, ok = runtime.Caller(0)
	line += 7 // adjust to match line of panic below
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// Initiate a sigpanic.
	var x *uintptr
	_ = *x
}
