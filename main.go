package main

import (
	"log"

	"github.com/scootpl/go-tensorflow-audio-example/app"
	"github.com/scootpl/go-tensorflow-audio-example/app/audio/wav"
)

const (
	batch                  = 600
	modelSpecificChunkSize = 150
	chunkSize              = modelSpecificChunkSize
)

func main() {
	c := app.NewCLI()

	if err := c.ParseArgs(); err != nil {
		log.Fatal(err)
	}

	reader, writer, err := c.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	w := wav.Init(reader, writer, batch, chunkSize)
	defer w.Close()

	s := app.New(c.ModelName(), w, batch, chunkSize)
	s.Process()
}
