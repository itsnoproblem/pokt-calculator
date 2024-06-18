package pocket

import (
	"fmt"
	"time"
)

const TypeClaim = "pocketcore/claim"
const TypeProof = "pocketcore/proof"

type Transaction struct {
	Hash          string
	Height        uint
	Time          time.Time
	Address       string
	Type          string
	ChainID       string
	NumRelays     uint
	PoktPerRelay  float64
	SessionHeight uint
	ExpireHeight  uint
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

// PoktAmount is calculated as:
// reward = NUM_RELAYS * RelaysToTokensMultiplier * ((FLOOR/ValidatorStakeFloorMultiplier)/( ServicerStakeWeightMultiplier*ValidatorStakeFloorMultiplier))^(ValidatorStakeFloorMultiplierExponent)
func (tx Transaction) PoktAmount() float64 {
	return tx.PoktPerRelay * float64(tx.NumRelays)
}
