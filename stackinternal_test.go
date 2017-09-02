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

	if got, want := c.file(), file; got != want {
		t.Errorf("got file == %v, want file == %v", got, want)
	}

	if got, want := c.line(), line; got != want {
		t.Errorf("got line == %v, want line == %v", got, want)
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
			// count frames to runtime.sigpanic
			panicIdx := -1
			for i, c := range Trace() {
				if c.name() == "runtime.sigpanic" {
					panicIdx = i
					break
				}
			}
			if panicIdx == -1 {
				t.Fatal("no runtime.sigpanic entry on the stack")
			}
			if got, want := Caller(panicIdx).name(), "runtime.sigpanic"; got != want {
				t.Errorf("sigpanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).name(), "github.com/go-stack/stack.TestCallerPanic"; got != want {
				t.Errorf("TestCallerPanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := Caller(panicIdx+1).line(), line; got != want {
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

type fholder struct {
	f func() CallStack
}

func (fh *fholder) labyrinth() CallStack {
	for {
		return fh.f()
	}
}

func TestTrace(t *testing.T) {
	t.Parallel()

	_, _, line, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	fh := fholder{
		f: func() CallStack {
			cs := Trace()
			return cs
		},
	}

	cs := fh.labyrinth()

	lines := []int{line + 7, line - 7, line + 12}

	for i, line := range lines {
		if got, want := cs[i].line(), line; got != want {
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
				if c.name() == "runtime.sigpanic" {
					panicIdx = i
					break
				}
			}
			if panicIdx == -1 {
				t.Fatal("no runtime.sigpanic entry on the stack")
			}
			if got, want := trace[panicIdx].name(), "runtime.sigpanic"; got != want {
				t.Errorf("sigpanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := trace[panicIdx+1].name(), "github.com/go-stack/stack.TestTracePanic"; got != want {
				t.Errorf("TestTracePanic frame: got name == %v, want name == %v", got, want)
			}
			if got, want := trace[panicIdx+1].line(), line; got != want {
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
