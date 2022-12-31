package filediff

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

type MagicNumber uint32

const (
	Blake2SigMagic MagicNumber = 0x72730137
)

type SignatureType struct {
	sigType    MagicNumber
	chunkSize  uint32
	strongLen  uint32
	strongSigs [][]byte
	weak2block map[uint32]int
}

func CalcStrongChecksum(data []byte) []byte {
	d := blake2b.Sum256(data)
	return d[:]
}

// Signature
// https://rsync.samba.org/tech_report/node1.html
// 1. Split data to the fixed-size chunks
// 2. Calculate rolling checksum and strong checksum and store them in binary and struct format
func Signature(input io.Reader, output io.Writer, chunkSize uint32) (*SignatureType, error) {
	err := binary.Write(output, binary.BigEndian, Blake2SigMagic)
	if err != nil {
		return nil, err
	}
	err = binary.Write(output, binary.BigEndian, chunkSize)
	if err != nil {
		return nil, err
	}
	err = binary.Write(output, binary.BigEndian, uint32(32))
	if err != nil {
		return nil, err
	}

	data := make([]byte, chunkSize)

	var ret SignatureType
	ret.weak2block = make(map[uint32]int)
	ret.sigType = Blake2SigMagic
	ret.strongLen = 32
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
		output.Write(strong)

		ret.weak2block[weak] = len(ret.strongSigs)
		ret.strongSigs = append(ret.strongSigs, strong)
	}

	return &ret, nil
}

// ReadSignature reads a signature from an io.Reader.
func ReadSignature(r io.Reader) (*SignatureType, error) {
	var magic MagicNumber
	err := binary.Read(r, binary.BigEndian, &magic)
	if err != nil {
		return nil, err
	}

	var chunkSize uint32
	err = binary.Read(r, binary.BigEndian, &chunkSize)
	if err != nil {
		return nil, err
	}

	var strongLen uint32
	err = binary.Read(r, binary.BigEndian, &strongLen)
	if err != nil {
		return nil, err
	}

	var strongSigs [][]byte
	weak2block := map[uint32]int{}

	for {
		var weakSum uint32
		err = binary.Read(r, binary.BigEndian, &weakSum)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		strongSum := make([]byte, strongLen)
		n, err := r.Read(strongSum)
		if err != nil {
			return nil, err
		}
		if n != int(strongLen) {
			return nil, fmt.Errorf("got only %d/%d bytes of the strong hash", n, strongLen)
		}

		weak2block[weakSum] = len(strongSigs)
		strongSigs = append(strongSigs, strongSum)
	}

	return &SignatureType{
		sigType:    magic,
		chunkSize:  chunkSize,
		strongLen:  strongLen,
		strongSigs: strongSigs,
		weak2block: weak2block,
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
