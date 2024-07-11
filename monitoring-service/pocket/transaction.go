package pocket

import (
	"fmt"
	sdk "monitoring-service/types"
	"time"
)

const TypeClaim = "pocketcore/claim"
const TypeProof = "pocketcore/proof"

type Transaction struct {
	Hash          string
	Height        int64
	Time          time.Time
	Address       string
	Type          string
	ChainID       string
	NumRelays     sdk.BigInt
	PoktPerRelay  sdk.BigDec
	SessionHeight uint
	ExpireHeight  int64
	AppPubkey     string
	ResultCode    int64
	IsConfirmed   bool
	Reward        Reward
}

func (tx Transaction) Chain() (Chain, error) {
	chain, err := ChainFromID(tx.ChainID)
	if err != nil {
		return Chain{}, fmt.Errorf("Transaction.Chain: %s", err)
	}

	return chain, nil
}
