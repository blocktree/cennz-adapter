package cennzTransaction


const calPeriod = 64

func GetEra(height uint64) []byte {
	//return []byte{}
	return []byte{0x0}

	//phase := height % calPeriod
	//
	//index := uint64(6)
	//trailingZero := index - 1
	//
	//var encoded uint64
	//if trailingZero > 1 {
	//	encoded = trailingZero
	//} else {
	//	encoded = 1
	//}
	//
	//if trailingZero < 15 {
	//	encoded = trailingZero
	//} else {
	//	encoded = 15
	//}
	//
	//encoded += phase / 1 << 4
	//
	//first := byte(encoded >> 8)
	//second := byte(encoded & 0xff)
	//
	//return []byte{second, first}
}
