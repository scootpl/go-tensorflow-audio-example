# Tensorflow Go API Audio Example

An example of using a neural network model (LSTM) with the Tensorflow Go API.

The model converts a clean guitar signal into an overdriven one, sonically similar to the Tube Screamer TS9

## Installation
Read documentation for installing the Go bindings for TensorFlow.

https://github.com/tensorflow/build/blob/master/golang_install_guide/README.md

## Input file
Input file can be in WAV format: 8, 16 or 24 bit (integer)

## Output file
Same as input

## Model
- model saved as: "saved_model.pb:
- input: 150 samples (float32)
- batch: multiple of 150 (default: 600)

## CLI Parameters


```
Usage of ./main:
  -i string
        input file name, "-" for Stdin
  -m string
        model directory
  -o string
        output file name
```

Model based on: https://github.com/GuitarML/GuitarLSTM