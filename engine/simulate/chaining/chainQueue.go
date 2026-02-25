package chaining

import (
	"github.com/warebuf/fd3/engine/chainTypes"
)

type Chain struct {
	stack     []chainTypes.ChainNode
	execution []chainTypes.ChainNode
}
