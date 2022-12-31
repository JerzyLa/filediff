package filediff

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allTestCases = []string{
	"000-blake2-512",
	"001-blake2-512",
	"002-blake2-512",
	"003-blake2-512",
	"004-blake2-512",
	"005-blake2-512",
	"006-blake2-2",
	"007-blake2-5",
	"007-blake2-4",
	"007-blake2-3",
	"008-blake2-512",
	"009-blake2-512",
	"010-blake2-512",
	"011-blake2-3",
}

func argsFromTestName(name string) (file string, windowLen uint32, err error) {
	segs := strings.Split(name, "-")
	if len(segs) != 3 {
		return "", 0, fmt.Errorf("invalid format for name %q", name)
	}

	file = segs[0]

	windowLen64, err := strconv.ParseInt(segs[2], 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("invalid window length %q", segs[2])
	}
	windowLen = uint32(windowLen64)

	return
}

type errorI interface {
	// testing#T.Error || testing#B.Error
	Error(args ...interface{})
}

func signature(t errorI, src io.Reader) *SignatureType {
	var (
		windowLen uint32 = 512
		bufSize          = 65536
	)

	s, err := Signature(
		bufio.NewReaderSize(src, bufSize),
		io.Discard,
		windowLen)
	if err != nil {
		t.Error(err)
	}

	return s
}

func TestSignature(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	for _, tt := range allTestCases {
		t.Run(tt, func(t *testing.T) {
			file, windowLen, err := argsFromTestName(tt)
			r.NoError(err)

			inputData, err := os.ReadFile("testdata/" + file + ".old")
			r.NoError(err)
			input := bytes.NewReader(inputData)

			output := &bytes.Buffer{}
			gotSig, err := Signature(input, output, windowLen)
			r.NoError(err)

			wantSig, err := ReadSignatureFile("testdata/" + tt + ".signature")
			r.NoError(err)
			a.Equal(wantSig.chunkSize, gotSig.chunkSize)
			a.Equal(wantSig.sigType, gotSig.sigType)
			a.Equal(wantSig.strongLen, gotSig.strongLen)

			outputData, err := io.ReadAll(output)
			r.NoError(err)
			expectedData, err := os.ReadFile("testdata/" + tt + ".signature")
			r.NoError(err)
			a.Equal(expectedData, outputData)
		})
	}
}
