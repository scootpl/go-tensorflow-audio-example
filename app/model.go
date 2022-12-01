package app

import (
	"bytes"
	"encoding/binary"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

type TS9 struct {
	savedModel *tf.SavedModel
	inputOp    *tf.Operation
	outputOp   *tf.Operation
	batch      int64
}

func NewTS9(dir string, batch int) (*TS9, error) {
	savedModel, err := tf.LoadSavedModel(dir, []string{"serve"}, nil)
	if err != nil {
		return nil, err
	}

	m := new(TS9)
	m.batch = int64(batch)
	m.savedModel = savedModel
	m.inputOp = m.savedModel.Graph.Operation("serving_default_conv1d_input")
	m.outputOp = m.savedModel.Graph.Operation("StatefulPartitionedCall")

	return m, nil
}

func (m *TS9) Inference(i []float32) ([]float32, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, i)
	r := bytes.NewReader(b.Bytes())

	tensor, err := tf.ReadTensor(tf.Float, []int64{m.batch, 150, 1}, r)
	if err != nil {
		return nil, err
	}

	in := map[tf.Output]*tf.Tensor{
		m.inputOp.Output(0): tensor,
	}

	out := []tf.Output{
		m.outputOp.Output(0),
	}

	output, err := m.savedModel.Session.Run(in, out, nil)
	if err != nil {
		return nil, err
	}

	value := output[0].Value().([][]float32)

	result := make([]float32, m.batch)
	for i := 0; i < int(m.batch); i++ {
		result[i] = value[i][0]
	}

	return result, nil
}
