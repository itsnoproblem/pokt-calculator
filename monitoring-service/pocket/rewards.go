package pocket

import sdk "monitoring-service/types"

const RewardScalingActivationHeight = 69243

type Reward struct {
	PoktAmount    sdk.BigDec
	NetPoktAmount sdk.BigDec
	StakeWeight   sdk.BigDec
	PoktPerRelay  sdk.BigDec
}

type MonthlyReward struct {
	Year                    uint
	Month                   uint
	TotalProofs             sdk.BigInt
	AvgSecsBetweenRewards   float64
	TotalSecsBetweenRewards float64
	DaysOfWeek              map[int]*DayOfWeek
	Transactions            []Transaction
}

type DayOfWeek struct {
	Name   string
	Proofs sdk.BigInt
}

func (r *MonthlyReward) PoktAmount() sdk.BigDec {
	var total = sdk.ZeroDec()
	for _, t := range r.Transactions {
		if t.IsConfirmed {
			total = total.Add(t.Reward.PoktAmount)
		}
	}
	return total
}
