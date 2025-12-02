package engine

type SummonType int

const (
	SummonTypeNormal SummonType = iota // can be summoned any time
	SummonTypeStart                    // summoned at start of the match
)

func CopySummonTypeSlice(src []SummonType) []SummonType {
	if src == nil {
		return nil
	}
	dst := make([]SummonType, len(src))
	copy(dst, src)
	return dst
}
