package filediff

type RollingChecksum struct {
	count  uint64
	s1, s2 uint16
}

const RollingChecksumCharOffset = 31

func NewRollingChecksum() RollingChecksum {
	return RollingChecksum{}
}

// CalcRollingChecksum is based on Mark Adler's adler-32 checksum
// The algorithm reference: https://rsync.samba.org/tech_report/node3.html
func CalcRollingChecksum(data []byte) uint32 {
	var sum RollingChecksum
	sum.Update(data)
	return sum.Digest()
}

func (r *RollingChecksum) Update(p []byte) {
	l := len(p)

	for n := 0; n < l; {
		if n+15 < l {
			for i := 0; i < 16; i++ {
				r.s1 += uint16(p[n+i])
				r.s2 += r.s1
			}
			n += 16
		} else {
			r.s1 += uint16(p[n])
			r.s2 += r.s1
			n += 1
		}
	}

	r.s1 += uint16(l * RollingChecksumCharOffset)
	r.s2 += uint16(((l * (l + 1)) / 2) * RollingChecksumCharOffset)
	r.count += uint64(l)
}

func (r *RollingChecksum) Rotate(out, in byte) {
	r.s1 += uint16(in) - uint16(out)
	r.s2 += r.s1 - uint16(r.count)*(uint16(out)+uint16(RollingChecksumCharOffset))
}

func (r *RollingChecksum) Rollin(in byte) {
	r.s1 += uint16(in) + uint16(RollingChecksumCharOffset)
	r.s2 += r.s1
	r.count += 1
}

func (r *RollingChecksum) Rollout(out byte) {
	r.s1 -= uint16(out) + uint16(RollingChecksumCharOffset)
	r.s2 -= uint16(r.count) * (uint16(out) + uint16(RollingChecksumCharOffset))
	r.count -= 1
}

func (r *RollingChecksum) Digest() uint32 {
	return (uint32(r.s2) << 16) | (uint32(r.s1) & 0xffff)
}

func (r *RollingChecksum) Reset() {
	r.count = 0
	r.s1 = 0
	r.s2 = 0
}
