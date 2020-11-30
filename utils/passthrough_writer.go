package utils

import "io"

type passThroughWriter struct {
	io.Writer
	destWriters []io.Writer
}

func NewPassThroughWriter(dest ...io.Writer) io.Writer {
	return &passThroughWriter{
		destWriters: dest,
	}
}

func (p *passThroughWriter) Write(data []byte) (int, error) {
	totalDataWritten := 0
	for i := range p.destWriters {
		curWriter := p.destWriters[i]
		var err error
		var currentWritten int
		if currentWritten, err = curWriter.Write(data); err != nil {
			panic(err) // todo should one write fail them all?
		}
		totalDataWritten += currentWritten
	}
	return totalDataWritten, nil
}
