package pocket

import (
	"fmt"
	sdk "monitoring-service/types"
)

type Params struct {
	RelaysToTokensMultiplier             sdk.BigInt
	ServicerStakeWeightMultiplier        sdk.BigDec
	ServicerStakeFloorMultiplier         sdk.BigInt
	ServicerStakeFloorMultiplierExponent sdk.BigDec
	ServicerStakeWeightCeiling           int64
	DaoAllocation                        sdk.BigInt
	ProposerPercentage                   sdk.BigInt
	ClaimExpirationBlocks                sdk.BigInt
}

type AllParams struct {
	AppParams    ParamGroup `json:"app_params"`
	AuthParams   ParamGroup `json:"auth_params"`
	GovParams    ParamGroup `json:"gov_params"`
	NodeParams   ParamGroup `json:"node_params"`
	PocketParams ParamGroup `json:"pocket_params"`
}

func (a AllParams) Validate() error {
	if a.AppParams == nil || len(a.AppParams) == 0 {
		return fmt.Errorf("app_params is empty")
	}

	if a.AuthParams == nil || len(a.AuthParams) == 0 {
		return fmt.Errorf("auth_params is empty")
	}

	if a.GovParams == nil || len(a.GovParams) == 0 {
		return fmt.Errorf("gov_params is empty")
	}

	if a.NodeParams == nil || len(a.NodeParams) == 0 {
		return fmt.Errorf("node_params is empty")
	}

	if a.PocketParams == nil || len(a.PocketParams) == 0 {
		return fmt.Errorf("pocket_params is empty")
	}

	return nil
}

type ParamGroup []Param

func (a ParamGroup) Get(k string) (string, bool) {
	for _, p := range a {
		if p.Key == k {
			return p.Value, true
		}
	}
	return "", false
}

type Param struct {
	Key   string `json:"param_key"`
	Value string `json:"param_value"`
}

const Pip22ExponentDenominator = 100

func (p Params) LegacyPoktPerRelay() sdk.BigDec {
	return (p.RelaysToTokensMultiplier.
		Quo(sdk.NewInt(1000000))).
		Mul(sdk.NewInt(100).Sub(p.DaoAllocation).Sub(p.ProposerPercentage).
			Quo(sdk.NewInt(Pip22ExponentDenominator))).
		ToDec()
}
