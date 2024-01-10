package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/google/uuid"
)

type OutputBuffer struct {
	sync.RWMutex
	data   []byte
	closed bool
	pty    *os.File
}

func (b *OutputBuffer) Write(d []byte) (n int, err error) {
	b.data = append(b.data, d...)
	return len(d), nil
}

func (b *OutputBuffer) SetPTY(pty *os.File) {
	b.Lock()
	defer b.Unlock()
	b.pty = pty
}

func (b *OutputBuffer) UnsetPTY() {
	b.Lock()
	defer b.Unlock()
	b.pty = nil
}

func (b *OutputBuffer) resize(size TTYSize) {
	b.RLock()
	defer b.RUnlock()
	if b.pty != nil {
		pty.Setsize(b.pty, &pty.Winsize{
			Rows: size.Rows,
			Cols: size.Cols,
			X:    size.X,
			Y:    size.Y,
		})
	}
}

func (b *OutputBuffer) Close() error {
	b.closed = true
	return nil
}

type TTY interface {
	io.ReadSeekCloser
	Resize(size TTYSize)
}

type OutputBufferSeeker struct {
	buf       *OutputBuffer
	offset    int64
	canResize func() bool
	closer    func()
}

func NewOutputBufferSeeker(buf *OutputBuffer, canResize func() bool, closer func()) *OutputBufferSeeker {
	return &OutputBufferSeeker{
		buf:       buf,
		canResize: canResize,
		closer:    closer,
	}
}

func (s *OutputBufferSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return 0, fmt.Errorf("seek whence not supported")
	}

	s.offset = offset
	return offset, nil
}

func (s *OutputBufferSeeker) Read(dst []byte) (n int, err error) {
	if s.offset >= int64(len(s.buf.data)) {
		if s.buf.closed {
			return 0, io.EOF
		}
		time.Sleep(200 * time.Millisecond)
		return 0, nil
	}

	n = copy(dst, s.buf.data[s.offset:])
	s.offset += int64(n)
	return n, nil
}

func (s *OutputBufferSeeker) Resize(size TTYSize) {
	if s.canResize() {
		s.buf.resize(size)
	}
}

func (s *OutputBufferSeeker) Close() error {
	s.closer()
	return nil
}

var _ TTY = (*OutputBufferSeeker)(nil)

type TTYSize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

type TTYOutput interface {
	Write(d []byte) (n int, err error)
	GetReadSeeker() TTY
}

type ttyOutputImpl struct {
	sync.RWMutex
	buffer  *OutputBuffer
	readers map[string]struct{}
}

func NewTTYOutput() TTYOutput {
	return &ttyOutputImpl{
		buffer:  &OutputBuffer{},
		readers: make(map[string]struct{}),
	}
}

func (o *ttyOutputImpl) Write(d []byte) (n int, err error) {
	return o.buffer.Write(d)
}

func (o *ttyOutputImpl) GetReadSeeker() TTY {
	id := uuid.NewString()
	o.Lock()
	defer o.Unlock()

	o.readers[id] = struct{}{}
	closeTTY := func() {
		o.Lock()
		defer o.Unlock()
		delete(o.readers, id)
	}
	canResize := func() bool {
		o.RLock()
		defer o.RUnlock()
		return len(o.readers) == 1
	}
	return NewOutputBufferSeeker(o.buffer, canResize, closeTTY)
}
