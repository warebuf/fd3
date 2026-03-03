package simulate

type Node interface {
	ID() int
}

type Stack struct {
	DeclarationIndex int64 // the latest effect gets the biggest index

	DeclarationTime int
	WindowTime      int
	ResolutionTime  int

	Effects []Node

	index int // position in arr
}

type PriorityQueue []*Stack
