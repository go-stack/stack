package stack

import (
	"runtime"
	"testing"
	"github.com/dkushner/stack"
	"strings"
	"strconv"
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

type fholder struct {
	f func() CallStack
}

func (fh *fholder) labyrinth() CallStack {
	for {
		return fh.f()
	}
	panic("this line only needed for go 1.0")
}

func TestTrace(t *testing.T) {
	t.Parallel()

	fh := fholder{
		f: func() CallStack {
			cs := Trace()
			return cs
		},
	}

	cs := fh.labyrinth()

	lines := []int{46, 36, 51}

	for i, line := range lines {
		if got, want := cs[i].line(), line; got != want {
			t.Errorf("got line[%d] == %v, want line[%d] == %v", i, got, i, want)
		}
	}
}

// Test stack handling originating from a sigpanic.
func TestTracePanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() != nil {
			trace := stack.Trace().TrimRuntime()

			if len(trace) != 6 {
				t.Errorf("got len(trace) == %v, want %v", len(trace), 6)
			}

			// Check frames in this file, the interceding frames are somewhat
			// platform-dependent.
			lines := []int64{68, 101}

			var local []int64
			for _, call := range trace {
				parts := strings.Split(call.String(), ":")
				if parts[0] == "stackinternal_test.go" {
					line, _ := strconv.ParseInt(parts[1], 10, 32)
					local = append(local, line)
				}
			}

			if len(local) != 2 {
				t.Errorf("expected %v local frames but got %v", 2, len(local))
			}

			for i, line := range lines {
				if got, want := local[i], line; got != want {
					t.Errorf("got line[%d] == %v, want line[%d] == %v", i, got, i, want)
				}
			}
		}
	}()

	// Initiate a sigpanic.
	var x *uintptr
	_ = *x
}
