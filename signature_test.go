package filediff

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allTestCases = []string{
	"000-chunksize-500",
	//"000-chunksize-512",
	//"001-chunksize-512",
	//"002-chunksize-512",
	//"003-chunksize-512",
	//"004-chunksize-512",
	//"005-chunksize-512",
	//"006-chunksize-2",
	//"007-chunksize-5",
	//"007-chunksize-4",
	//"007-chunksize-3",
	//"008-chunksize-512",
	//"009-chunksize-512",
	//"010-chunksize-512",
	//"011-chunksize-3",
}

func TestSignature(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	for _, tt := range allTestCases {
		t.Run(tt, func(t *testing.T) {
			segs := strings.Split(tt, "-")
			file := segs[0]
			chunkSize, err := strconv.ParseInt(segs[2], 10, 32)
			r.NoError(err)

			inputData, err := os.ReadFile("testdata/" + file + ".old")
			r.NoError(err)
			input := bytes.NewReader(inputData)

			output := &bytes.Buffer{}
			gotSig, err := Signature(input, output, uint32(chunkSize))
			r.NoError(err)

			wantSig, err := ReadSignatureFile("testdata/" + tt + ".signature")
			r.NoError(err)
			a.Equal(wantSig.chunkSize, gotSig.chunkSize)

			outputData, err := io.ReadAll(output)
			r.NoError(err)
			expectedData, err := os.ReadFile("testdata/" + tt + ".signature")
			r.NoError(err)
			a.Equal(expectedData, outputData)
		})
	}
}
