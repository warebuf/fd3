package copyslice

func CopySliceUint(src []uint) []uint {
	if src == nil {
		return nil
	}
	dst := make([]uint, len(src))
	copy(dst, src)
	return dst
}

func CopySliceInt(src []int) []int {
	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}
