package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
	"unsafe"
)

type ModelMock struct{}

func (m *ModelMock) Inference(i []float32) ([]float32, error) {
	return i, nil
}

type AudioControlMock struct {
	source    []byte
	buf       *bytes.Buffer
	w         []float32
	chunkSize int
}

func (a *AudioControlMock) Init(bufSize int) {
	a.source = make([]byte, bufSize)
	a.buf = bytes.NewBuffer(a.source)
	a.buf.Reset()

	floatSize := int(unsafe.Sizeof(float32(1.0)))
	f := make([]float32, bufSize/floatSize)
	for i := 0; i < len(f); i++ {
		f[i] = float32(i)
	}

	binary.Write(a.buf, binary.LittleEndian, f)
}

func (a *AudioControlMock) Load(samples int) ([]float32, int, error) {
	floatSize := int(unsafe.Sizeof(float32(1.0)))
	b := make([]byte, samples*floatSize)
	n, err := a.buf.Read(b)
	if err != nil || n == 0 {
		return nil, 0, err
	}

	r := bytes.NewReader(b)
	out := make([]float32, samples)
	err = binary.Read(r, binary.LittleEndian, out)
	if err != nil {
		return nil, 0, err
	}

	return out, n / floatSize, nil
}

func (a *AudioControlMock) Write(w []float32) error {
	fmt.Printf("chunk: %v\n", w)
	a.w = append(a.w, w...)
	return nil
}

func (a *AudioControlMock) Close() {
}

func TestApp_Process(t *testing.T) {

	tests := []struct {
		name                  string
		chunkSize             int
		testProbeSizeInChunks int
		batch                 int
		want                  string
	}{
		{
			name:                  "test1",
			chunkSize:             5,
			testProbeSizeInChunks: 12,
			batch:                 10,
			want:                  "5EC85DEBADA8AFA0626F320B1A8F0DFAA3652258842E9DE6E06F15E65410E4A7",
		},
		{
			name:                  "test2",
			chunkSize:             5,
			testProbeSizeInChunks: 9,
			batch:                 10,
			want:                  "22E0F504880117E70CF2FB3B6DFE1A09D8E7A6EEB065B81BCBAEE8A509E13BC6",
		},
		{
			name:                  "test3",
			chunkSize:             5,
			testProbeSizeInChunks: 10,
			batch:                 10,
			want:                  "E721E7F492E0340E7D3B2309D11DB5CC160E1A485B920FFCCD3CACCF13E3B0C6",
		},
		{
			name:                  "test4",
			chunkSize:             5,
			testProbeSizeInChunks: 3,
			batch:                 10,
			want:                  "262C9FE970B90652F2101BA2CF6D0DF4D7602EDEB1708D0E670E788D7BEA9B00",
		},
	}

	for _, test := range tests {
		floatSize := int(unsafe.Sizeof(float32(1.0)))
		marginSizeInBytes := (test.chunkSize * floatSize) - floatSize
		bufSize := (test.testProbeSizeInChunks * test.chunkSize * floatSize) + marginSizeInBytes

		ac := new(AudioControlMock)
		ac.Init(bufSize)

		ac.chunkSize = test.chunkSize
		mm := new(ModelMock)

		t.Run("Process() test", func(t *testing.T) {
			a := &App{
				audio:     ac,
				model:     mm,
				batch:     test.batch,
				chunkSize: test.chunkSize,
				buffLen:   ((test.batch / test.chunkSize) + 1) * test.chunkSize,
			}
			a.Process()

			s := sha256.New()
			binary.Write(s, binary.LittleEndian, ac.w)
			fmt.Printf("Hash: %X\n", s.Sum(nil))

			if fmt.Sprintf("%X", s.Sum(nil)) != test.want {
				t.Errorf("Test: %s wrong checksum (got: %X)\n", test.name, s.Sum(nil))
			}

		})
	}
}
