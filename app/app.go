package app

import (
	"io"
	"log"
)

type App struct {
	audio     AudioFileController
	model     Model
	batch     int
	chunkSize int
	buffer    []float32
	buffLen   int
}

type AudioFileController interface {
	Load(samples int) ([]float32, int, error)
	Write([]float32) error
	Close()
}

type Model interface {
	Inference(i []float32) ([]float32, error)
}

func New(name string, audio AudioFileController, batch, chunkSize int) *App {
	m, err := NewTS9(name, batch)
	if err != nil {
		log.Fatal(err)
	}

	return &App{
		audio:     audio,
		model:     m,
		batch:     batch,
		chunkSize: chunkSize,
		buffLen:   ((batch / chunkSize) + 1) * chunkSize,
	}
}

// Copy last chunk to beginning of the buffer
func (a *App) copyLastChunk() {
	copy(a.buffer[:a.chunkSize], a.buffer[len(a.buffer)-a.chunkSize:])
}

// Load chunks at position
//
// Params
// 'pos': start position in chunks, 0 is first chunk
//
// Results
// 'n': read samples
// 'loaded': is buffer fully loaded?
// 'err': error or nil
func (a *App) loadChunksAtPosition(pos int) (n int, loaded bool, err error) {
	buf, n, err := a.audio.Load(a.buffLen - (pos * a.chunkSize))
	if err != nil || n == 0 {
		return
	}

	if n == a.buffLen {
		loaded = true
	}

	copy(a.buffer[pos*a.chunkSize:], buf[:n])
	a.buffer = a.buffer[:n+(pos*a.chunkSize)]
	return
}

func (a *App) initBuffer() {
	a.buffer = make([]float32, a.buffLen)
}

func (a *App) Process() {
	var (
		n           int
		loaded      bool
		exit        bool
		batchBuffer = make([]float32, 0, (a.buffLen-a.chunkSize)*a.chunkSize)
	)

	a.initBuffer()

	// Load chunks starting at pos 0
	n, loaded, err := a.loadChunksAtPosition(0)
	if err != nil || n == 0 {
		return
	}

	for {
		if !loaded {
			// Buffer is not fully loaded
			// We should prepare complete buffer to run batch
			tmp := make([]float32, a.buffLen-len(a.buffer))
			a.buffer = append(a.buffer, tmp...)

			// there is no data in last chunk
			if n <= (a.buffLen - a.chunkSize - 1) {
				exit = true
			}
		}

		batchBuffer = batchBuffer[:0]

		for i := 0; i < a.batch; i += 1 {
			batchBuffer = append(batchBuffer, a.buffer[i:i+a.chunkSize]...)
		}

		modelout, err := a.model.Inference(batchBuffer)
		if err != nil {
			log.Fatal(err)
		}

		if err := a.audio.Write(modelout); err != nil {
			log.Println("Write: ", err)
			return
		}

		// Copy last chunk to beginning
		a.copyLastChunk()

		// Load n-1 chunks, starts from 1 (second chunk)
		n, loaded, err = a.loadChunksAtPosition(1)
		if err != io.EOF && err != nil || exit {
			return
		}
	}
}
