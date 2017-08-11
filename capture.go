package sweet

import (
	"bytes"
	"os"
)

type pipeCapture struct {
	r   *os.File
	w   *os.File
	buf bytes.Buffer

	runChan  chan struct{}
	stopChan chan struct{}
}

func newPipeCapture() (*pipeCapture, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	pc := &pipeCapture{
		r: r,
		w: w,

		runChan:  make(chan struct{}),
		stopChan: make(chan struct{}),
	}

	go pc.readPipe()
	<-pc.runChan

	return pc, nil
}

func (pc *pipeCapture) readPipe() {
	close(pc.runChan)
	buf := make([]byte, 2048)

readLoop:
	for {
		select {
		case <-pc.stopChan:
			break readLoop
		default:
		}

		rN, err := pc.r.Read(buf)
		if err != nil {
			return
		}

		writeStart := 0

		for {
			wN, err := pc.buf.Write(buf[writeStart : rN-writeStart])
			if err != nil {
				return
			}
			writeStart += wN
			if writeStart == rN {
				break
			}
		}
	}
}

func (pc *pipeCapture) W() *os.File {
	return pc.w
}

func (pc *pipeCapture) Buffer() []byte {
	return pc.buf.Bytes()
}

func (pc *pipeCapture) Close() error {
	err := pc.w.Close()
	if err != nil {
		return err
	}
	err = pc.r.Close()
	if err != nil {
		return err
	}

	return nil
}
