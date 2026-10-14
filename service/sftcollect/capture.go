package sftcollect

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

const maxCaptureBytes = 512 << 10

// Capture tees the client-facing response into a bounded buffer.
// Writes always go to the real gin writer first; overflow only drops collection.
type Capture struct {
	gin.ResponseWriter
	body     *bytes.Buffer
	overflow bool
}

func Attach(c *gin.Context) *Capture {
	if c == nil || c.Writer == nil {
		return nil
	}
	cap := &Capture{
		ResponseWriter: c.Writer,
		body:           bytes.NewBuffer(make([]byte, 0, 4096)),
	}
	c.Writer = cap
	return cap
}

func (w *Capture) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if n > 0 {
		w.note(b[:n])
	}
	return n, err
}

func (w *Capture) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *Capture) Overflow() bool {
	return w != nil && w.overflow
}

func (w *Capture) StatusCode() int {
	if w == nil {
		return 0
	}
	return w.ResponseWriter.Status()
}

func (w *Capture) CopyBody() []byte {
	if w == nil || w.body == nil || w.overflow {
		return nil
	}
	src := w.body.Bytes()
	out := make([]byte, len(src))
	copy(out, src)
	return out
}

func (w *Capture) note(b []byte) {
	if w.overflow || w.body == nil {
		return
	}
	if w.body.Len()+len(b) > maxCaptureBytes {
		w.overflow = true
		return
	}
	_, _ = w.body.Write(b)
}
