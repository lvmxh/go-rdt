package util

// cbm are all consecutive bits
var HexMap = map[byte]int{
	49:  1,
	51:  2,
	55:  3,
	102: 4,
}

func CbmLen(cbm string) int {
	len := 0
	for i, _ := range cbm {
		len += HexMap[cbm[i]]
	}
	return len
}
