package wav

import (
	"io"
	"math"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Wav struct {
	enc        *wav.Encoder
	wav        *wav.Decoder
	writer     io.WriteSeeker
	intBuffer  *audio.IntBuffer
	firstChunk bool
}

func Init(r io.ReadSeeker, w io.WriteSeeker, batch, chunkSize int) *Wav {
	a := new(Wav)
	a.firstChunk = true
	a.writer = w
	a.wav = wav.NewDecoder(r)

	a.intBuffer = &audio.IntBuffer{
		Data: make([]int, ((batch/chunkSize)+1)*chunkSize),
	}

	return a
}

func (w *Wav) Close() {
	w.enc.Close()
}

func (w *Wav) Write(b []float32) error {
	out := w.float32ToIntBuffer(b)
	return w.enc.Write(out)
}

func (w *Wav) Load(samples int) ([]float32, int, error) {
	w.intBuffer.Data = w.intBuffer.Data[:samples]

	n, err := w.wav.PCMBuffer(w.intBuffer)
	if err != nil || n == 0 {
		return nil, 0, err
	}

	if w.firstChunk {
		w.writeHeader()
		w.firstChunk = false
	}

	return w.intBuffer.AsFloat32Buffer().Data, n, nil
}

func (w *Wav) writeHeader() {
	w.enc = wav.NewEncoder(w.writer,
		int(w.wav.SampleRate),
		int(w.wav.BitDepth),
		int(w.wav.NumChans),
		int(w.wav.WavAudioFormat))
}

func (w *Wav) float32ToIntBuffer(in []float32) *audio.IntBuffer {
	b := &audio.IntBuffer{
		Format:         w.wav.Format(),
		Data:           make([]int, len(in)),
		SourceBitDepth: int(w.wav.BitDepth),
	}

	factor := float32(math.Pow(2, float64(w.wav.BitDepth)-1))
	for i, f := range in {
		b.Data[i] = int(f * factor)
	}

	return b
}
