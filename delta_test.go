package filediff

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func TestDelta(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	for _, tt := range allTestCases {
		t.Run(tt, func(t *testing.T) {
			file, _, err := argsFromTestName(tt)
			r.NoError(err)

			sig, err := ReadSignatureFile("testdata/" + tt + ".signature")
			r.NoError(err)

			newFile, err := os.Open("testdata/" + file + ".new")
			r.NoError(err)

			deltaBuffer := &bytes.Buffer{}
			err = Delta(sig, newFile, deltaBuffer)
			r.NoError(err)
			delta, err := io.ReadAll(deltaBuffer)
			r.NoError(err)

			expectedDelta, err := os.ReadFile("testdata/" + tt + ".delta")
			r.NoError(err)

			a.Equal(expectedDelta, delta)
		})
	}
}
