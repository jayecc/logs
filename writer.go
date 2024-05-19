package logs

import (
	"errors"
	"golang.org/x/sync/errgroup"
	"io"
	"sync/atomic"
)

type Writer struct {
	ch chan []byte    // 用于存储要写入的数据的通道
	cl atomic.Bool    // 用于标识通道是否已关闭的原子变量
	g  errgroup.Group // 用于管理协程的协程组
	w  io.WriteCloser // 写入的目标
}

func NewWriter(w io.WriteCloser, bufferSize int) *Writer {

	fw := &Writer{
		ch: make(chan []byte, bufferSize),
		cl: atomic.Bool{},
		g:  errgroup.Group{},
		w:  w,
	}

	fw.g.Go(fw.handle)

	return fw
}

func (w *Writer) handle() (err error) {
	for {
		select {
		case d, ok := <-w.ch:
			if !ok {
				return
			}
			if _, err = w.w.Write(d); err != nil {
				return err
			}
		}
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if p == nil {
		return
	}
	if w.cl.Load() {
		return 0, errors.New("channel is closed") // 如果通道已关闭，则返回错误信息
	}
	s := string(p)
	select {
	case w.ch <- []byte(s): // 将数据写入通道中
		return len(p), nil
	default:
		return 0, errors.New("channel is full")
	}
}

func (w *Writer) Close() error {
	w.g.Go(w.w.Close)
	w.cl.Store(true)
	close(w.ch)
	return w.g.Wait()
}
