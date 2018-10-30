package main

func s16le(hi, lo byte) int16 {
	sign := hi & (1 << 7)
	v := int16(hi&0x7f)<<8 | int16(lo)
	if sign != 0 {
		v = -0x8000 + v
	}
	return v
}

func mix(a, b int16) int16 {
	v := (float32(a) + float32(b)) / 0x7fff
	if v <= -1.25 {
		v = -0.984375
	} else if v >= 1.25 {
		v = 0.984375
	} else {
		v = 1.1*v - 0.2*v*v*v
	}
	return int16(v * 0x7fff)
}
