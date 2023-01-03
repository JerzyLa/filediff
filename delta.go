package filediff

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/smallnest/ringbuffer"
	"io"
)

const (
	// OpCopy copy data operation
	// Used to copy bytes from the old data
	OpCopy uint8 = 69
	// OpLiteral literal data operation
	// Used to add new bytes to the old data
	OpLiteral uint8 = 65
)

func Delta(signature *SignatureType, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	chunk := ringbuffer.New(int(signature.chunkSize))

	rollingChecksum := NewRollingChecksum()
	previousByte := byte(0)
	res := newDeltaResult(output)

	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if chunk.IsFull() {
			previousByte, err = chunk.ReadByte()
			if err != nil {
				return err
			}
			rollingChecksum.Rollout(previousByte)
			// Add operation literal to the result
			err = res.addOpLiteral(previousByte)
			if err != nil {
				return err
			}
		}

		err = chunk.WriteByte(b)
		if err != nil {
			return err
		}

		rollingChecksum.Rollin(b)

		if !chunk.IsFull() {
			continue
		}

		// Check if the chunk is found in the signature
		if chunkIndex, ok := signature.rolling2chunk[rollingChecksum.Digest()]; ok {
			strongChecksum := CalcStrongChecksum(chunk.Bytes())
			if bytes.Equal(signature.strongChecksums[chunkIndex], strongChecksum) {
				// Reset the chunk and rolling hash if match was found
				chunk.Reset()
				rollingChecksum.Reset()
				// Add operation copy to the result
				err = res.addOpCopy(uint64(chunkIndex)*uint64(signature.chunkSize), uint64(signature.chunkSize))
				if err != nil {
					return err
				}
			}
		}
	}

	for _, b := range chunk.Bytes() {
		// Add operation literal to the result
		err := res.addOpLiteral(b)
		if err != nil {
			return err
		}
	}

	err := res.write()
	if err != nil {
		return err
	}

	return nil
}

// deltaResult stores and updates delta results
type deltaResult struct {
	op     uint8
	pos    uint64
	len    uint64
	data   []byte
	output io.Writer
}

func newDeltaResult(output io.Writer) deltaResult {
	return deltaResult{
		output: output,
	}
}

// addOpCopy adds operation copy to the delta result
func (r *deltaResult) addOpCopy(pos uint64, len uint64) error {
	if r.op != OpCopy {
		err := r.write()
		if err != nil {
			return err
		}
		r.op = OpCopy
		r.pos = pos
	}
	r.len += len
	return nil
}

// addOpLiteral adds operation literal to the delta result
func (r *deltaResult) addOpLiteral(b byte) error {
	if r.op != OpLiteral {
		err := r.write()
		if err != nil {
			return err
		}
		r.op = OpLiteral
	}
	r.data = append(r.data, b)
	r.len += 1
	return nil
}

// write writes the delta result to the given output
func (r *deltaResult) write() error {
	switch r.op {
	case OpCopy:
		cmd := OpCopy
		switch intSize(r.pos) {
		case 2:
			cmd += 4
		case 4:
			cmd += 8
		case 8:
			cmd += 12
		}
		switch intSize(r.len) {
		case 2:
			cmd += 1
		case 4:
			cmd += 2
		case 8:
			cmd += 3
		}
		err := binary.Write(r.output, binary.BigEndian, cmd)
		if err != nil {
			return err
		}
		err = write(r.output, r.pos)
		if err != nil {
			return err
		}
		err = write(r.output, r.len)
		if err != nil {
			return err
		}

	case OpLiteral:
		cmd := OpLiteral
		switch intSize(r.len) {
		case 2:
			cmd += 1
		case 4:
			cmd += 2
		case 8:
			cmd += 3
		}
		err := binary.Write(r.output, binary.BigEndian, cmd)
		if err != nil {
			return err
		}
		err = write(r.output, r.len)
		if err != nil {
			return err
		}
		_, err = r.output.Write(r.data)
		if err != nil {
			return err
		}
		r.data = r.data[:0]
	}
	r.pos = 0
	r.len = 0
	return nil
}

// write writes data to the buffer on the minimum number of bytes needed
func write(writer io.Writer, data uint64) error {
	switch intSize(data) {
	case 1:
		return binary.Write(writer, binary.BigEndian, uint8(data))
	case 2:
		return binary.Write(writer, binary.BigEndian, uint16(data))
	case 4:
		return binary.Write(writer, binary.BigEndian, uint32(data))
	case 8:
		return binary.Write(writer, binary.BigEndian, data)
	}
	return fmt.Errorf("invalid size: %v", intSize(data))
}

func intSize(d uint64) uint8 {
	switch {
	case d == uint64(uint8(d)):
		return 1
	case d == uint64(uint16(d)):
		return 2
	case d == uint64(uint32(d)):
		return 4
	default:
		return 8
	}
}
