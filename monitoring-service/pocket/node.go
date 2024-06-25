package pocket

import "time"

type Node struct {
	Address           string
	Pubkey            string
	Balance           uint
	StakedBalance     uint
	ServiceURL        string
	IsJailed          bool
	Chains            []Chain
	IsSynced          bool
	LatestBlockHeight uint
	LatestBlockTime   time.Time
	CustodialNode     bool
	OutputAddress     string
}

type Session struct {
	ChainID      string
	Height       uint
	AppPublicKey string
	NumRelays    uint
}
