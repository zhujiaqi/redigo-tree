package redigotree

// Returns a new slice containing all strings in the slice that satisfy the predicate f.
// Borrowed from https://play.golang.org/p/3PNdke3Wia
func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Borrowed from https://stackoverflow.com/questions/28848187/golang-convert-int8-to-string
func B2S(bs []uint8) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}
