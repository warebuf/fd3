package chainTypes

// every chain node has a parent too

type ChainNode interface {
	Kind() string
}
