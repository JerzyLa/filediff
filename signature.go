package filediff

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

// SignatureType stores rolling checksums and strong checksums
type SignatureType struct {
	chunkSize       uint32
	strongChecksums [][]byte
	rolling2chunk   map[uint32]int
}

// CalcStrongChecksum calculates BLAKE2b-256 checksum
func CalcStrongChecksum(data []byte) []byte {
	d := blake2b.Sum256(data)
	return d[:]
}

// Signature calculates a signature from the given data
// The algorithm reference: https://rsync.samba.org/tech_report/node2.html
// 1. Split data to the fixed-size chunks
// 2. Calculate rolling checksum and strong checksum
// 3. Store checksums in the output and return the signature struct
func Signature(input io.Reader, output io.Writer, chunkSize uint32) (*SignatureType, error) {
	err := binary.Write(output, binary.BigEndian, chunkSize)
	if err != nil {
		return nil, err
	}

	data := make([]byte, chunkSize)

	var ret SignatureType
	ret.rolling2chunk = make(map[uint32]int)
	ret.chunkSize = chunkSize

	for {
		n, err := input.Read(data)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		chunk := data[:n]

		weak := CalcRollingChecksum(chunk)
		err = binary.Write(output, binary.BigEndian, weak)
		if err != nil {
			return nil, err
		}

		strong := CalcStrongChecksum(chunk)
		_, err = output.Write(strong)
		if err != nil {
			return nil, err
		}

		ret.rolling2chunk[weak] = len(ret.strongChecksums)
		ret.strongChecksums = append(ret.strongChecksums, strong)
	}

	return &ret, nil
}

// ReadSignature reads a signature from an io.Reader.
func ReadSignature(r io.Reader) (*SignatureType, error) {
	var chunkSize uint32
	err := binary.Read(r, binary.BigEndian, &chunkSize)
	if err != nil {
		return nil, err
	}

	var strongChecksums [][]byte
	rolling2chunk := map[uint32]int{}

	for {
		var rollingChecksum uint32
		err = binary.Read(r, binary.BigEndian, &rollingChecksum)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		strongChecksum := make([]byte, 32)
		n, err := r.Read(strongChecksum)
		if err != nil {
			return nil, err
		}
		if n != 32 {
			return nil, fmt.Errorf("got only %d/%d bytes of the strong hash", n, 32)
		}

		rolling2chunk[rollingChecksum] = len(strongChecksums)
		strongChecksums = append(strongChecksums, strongChecksum)
	}

	return &SignatureType{
		chunkSize:       chunkSize,
		strongChecksums: strongChecksums,
		rolling2chunk:   rolling2chunk,
	}, nil
}

// ReadSignatureFile reads a signature from the file at path.
func ReadSignatureFile(path string) (*SignatureType, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadSignature(f)
}
