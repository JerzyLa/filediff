package filediff

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
)

var testCases = []string{
	/// 000.old and 000.new files are the same (5000 bytes) and chunk size is 500 bytes
	/// The delta is OP_COPY 0 5000
	"000-chunksize-500",
	/// 000.old and 000.new files are the same (5000 bytes) and chunk size is 512 bytes
	/// The delta is OP_COPY 0 4608 (9 chunks) and OP_LITERAL 392
	"000-chunksize-512",
	/// 001.new has added 300 bytes at the end of the file 001.old and chunk size is 512 bytes
	/// The delta is OP_COPY 0 6656 (13 chunks) and OP_LITERAL 644
	"001-chunksize-512",
	/// 002.new has added 1000 bytes at the beginning of the file 002.old and chunk size is 512 bytes
	/// The delta is OP_LITERAL 1000 OP_COPY 0 6656 (13 chunks) OP_LITERAL 344 (left bytes)
	"002-chunksize-512",
	/// 003.new has added bytes in the middle of the file 003.old and chunk size is 512 bytes
	"003-chunksize-512",
	/// 004.new has modified bytes in the middle of the 004.old file and chunk size is 512
	"004-chunksize-512",
	/// 005.new has added bytes at the beginning and the end of the 005.old and chunk size is 512
	"005-chunksize-512",
	/// 006.new has added one byte in the middle of the file 006.old and chunk size is 2
	/// The delta is OP_COPY 0 2 (1 chunk) and OP_LITERAL 2
	"006-chunksize-2",
	/// 007.new has removed bytes form file 007.old
	"007-chunksize-5",
	"007-chunksize-4",
	"007-chunksize-3",
	/// 008.new is empty file
	/// The delta is empty
	"008-chunksize-512",
	/// 009.new contains only added bytes to the empty 009.old file
	"009-chunksize-512",
	/// 010.new and 010.old are both empty files
	/// The delta is empty
	"010-chunksize-512",
	/// 011.new has 3 beginning bytes removed from 011.old
	"011-chunksize-3",
}

func TestDelta(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt, func(t *testing.T) {
			segs := strings.Split(tt, "-")
			file := segs[0]

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
