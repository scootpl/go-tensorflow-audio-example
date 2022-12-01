package app

import (
	"errors"
	"flag"
	"io"
	"os"
)

type Config struct {
	input, output     string
	inputFd, outputFd *os.File
	model             string
}

func NewCLI() *Config {
	return new(Config)
}

func (c *Config) ParseArgs() error {
	input := flag.String("i", "", "input file name, \"-\" for Stdin")
	output := flag.String("o", "", "output file name")
	model := flag.String("m", "", "model directory")

	flag.Parse()

	if *input == "" || *output == "" || *model == "" {
		return errors.New("try -h for help")
	}

	c.input = *input
	c.output = *output
	c.model = *model
	return nil
}

func (c *Config) ModelName() string {
	return c.model
}

func (c *Config) Init() (io.ReadSeeker, io.WriteSeeker, error) {
	var (
		reader io.ReadSeeker
		writer io.WriteSeeker
		err    error
	)

	if c.input == "-" {
		reader = os.Stdin
	} else {
		c.inputFd, err = os.Open(c.input)
		if err != nil {
			return nil, nil, err
		}
		reader = c.inputFd
	}

	c.outputFd, err = os.Create(c.output)
	if err != nil {
		return nil, nil, err
	}
	writer = c.outputFd

	return reader, writer, nil
}

func (c *Config) Close() {
	c.outputFd.Close()

	if c.input != "-" {
		c.inputFd.Close()
	}
}
