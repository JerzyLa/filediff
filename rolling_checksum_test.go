package filediff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRollingChecksumBrandNew(t *testing.T) {
	a := assert.New(t)
	r := RollingChecksum{}

	a.Equal(uint32(0x00000000), r.Digest())
}

func TestRollingChecksumRollinRollout(t *testing.T) {
	a := assert.New(t)
	r := RollingChecksum{}

	r.Rollin(222)
	a.Equal(uint32(0x00FD00FD), r.Digest())
	r.Rollin(11)
	a.Equal(uint32(0x02240127), r.Digest())
	r.Rollin(0)
	a.Equal(uint32(0x036A0146), r.Digest())
	r.Rollin(13)
	a.Equal(uint32(0x04DC0172), r.Digest())
	r.Rollin(7)
	a.Equal(uint32(0x06740198), r.Digest())

	r.Rollout(222)
	a.Equal(uint32(0x0183009B), r.Digest())
	r.Rollout(11)
	a.Equal(uint32(0x00DB0071), r.Digest())
	r.Rollout(0)
	a.Equal(uint32(0x007E0052), r.Digest())

	r.Rollin(1)
	a.Equal(uint32(0x00F00072), r.Digest())
}

func TestRollingChecksumUpdate(t *testing.T) {
	a := assert.New(t)
	r := RollingChecksum{}
	data := []byte{222, 11, 0, 13, 7}
	moreData := []byte{66, 171, 8}

	r.Update(data)
	a.Equal(uint32(0x06740198), r.Digest())
	r.Update(moreData)
	a.Equal(uint32(0x0E1A02EA), r.Digest())
}

func TestRollingChecksumRotate(t *testing.T) {
	a := assert.New(t)
	r := RollingChecksum{}
	data := []byte{222, 11, 0, 13, 7}

	r.Update(data)

	r.Rotate(222, 39)
	a.Equal(uint32(0x026400E1), r.Digest())
	r.Rotate(11, 177)
	a.Equal(uint32(0x03190187), r.Digest())
	r.Rotate(0, 0)
	a.Equal(uint32(0x04050187), r.Digest())
}

// Sanity check to verify that feeding all the data at once (with Update())
// gives the same result as feeding data one byte at a time (with Rotate(),
// Rollin(), and Rollout()).
func TestRollingChecksumConsistency(t *testing.T) {
	a := assert.New(t)
	data1 := []byte{ /* */ 66, 1, 111, 54, 171, 12, 255, 199, 1, 2, 7, 12, 54, 43, 101}
	data2 := []byte{4, 22, 66, 1, 111, 54, 171, 12, 255, 199, 1, 2, 7, 12, 54 /*    */}

	rk1 := RollingChecksum{}
	rk1.Update(data1)

	rk2 := RollingChecksum{}
	for _, v := range data2 {
		rk2.Rollin(v)
	}
	rk2.Rotate(4, 43)
	rk2.Rollout(22)
	rk2.Rollin(101)

	a.Equal(rk1.Digest(), rk2.Digest())
}

// This used to trigger a bug in Rotate(), caused by the use of subtraction with
// bytes, which should be using proper signed integers instead.
func TestRollingChecksumRotateByteSubtractionBug(t *testing.T) {
	a := assert.New(t)

	rk1 := RollingChecksum{}
	rk1.Rollin(1) // { 1 }
	a.Equal(uint32(0x00200020), rk1.Digest())

	rk2 := RollingChecksum{}
	rk2.Rollin(2)    // { 2 }
	rk2.Rotate(2, 1) // { 1 }

	a.Equal(rk1.Digest(), rk2.Digest())
}
