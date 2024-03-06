package types

// PoolIDForDenomPair defines a structure to associate a pool ID
// to swap between denoms A and B.
type PoolIDForDenomPair struct {
	DenomA string
	DenomB string
	PoolID uint64
}
