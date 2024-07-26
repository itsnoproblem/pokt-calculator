package monitoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"monitoring-service/pocket"
	pocketnode "monitoring-service/provider/pocket"
	sdk "monitoring-service/types"
)

type PocketProvider interface {
	NodeProvider(address string) (pocketnode.Provider, error)
	SimulateRelay(servicerUrl, chainID string, payload json.RawMessage) (json.RawMessage, error)
	AccountTransactions(address string, page uint, perPage uint, sort string) ([]pocket.Transaction, error)
	Transaction(hash string) (pocket.Transaction, error)
	BlockTime(height int64) (time.Time, error)
	Node(address string) (pocket.Node, error)
	NodeAtHeight(address string, height int64) (pocket.Node, error)
	Balance(address string) (uint, error)
	Param(name string, height int64) (string, error)
	AllParams(height int64, forceRefresh bool) (pocket.AllParams, error)
	Height() (uint, error)
}

func NewService(provider PocketProvider) Service {
	return Service{
		provider: provider,
	}
}

type Service struct {
	provider PocketProvider
}

func (s *Service) Height() (uint, error) {
	height, err := s.provider.Height()
	if err != nil {
		return 0, fmt.Errorf("Height: %s", err)
	}

	return height, nil
}

func (s *Service) Transaction(hash string) (pocket.Transaction, error) {
	txn, err := s.provider.Transaction(hash)
	if err != nil {
		return pocket.Transaction{}, fmt.Errorf("Transaction: %s", err)
	}

	txn.Time, err = s.provider.BlockTime(txn.Height)
	if err != nil {
		return pocket.Transaction{}, fmt.Errorf("Transaction: %s", err)
	}

	return txn, nil
}

func (s *Service) BlockTimes(heights []int64) (map[int64]time.Time, error) {
	times := make(map[int64]time.Time, len(heights))
	for _, id := range heights {
		var err error
		if times[id], err = s.provider.BlockTime(id); err != nil {
			return nil, fmt.Errorf("BlockTimes: %s", err)
		}
	}

	return times, nil
}

func (s *Service) ParamsAtHeight(height int64, forceRefresh bool) (pocket.Params, error) {
	params := pocket.Params{}

	allParams, err := s.provider.AllParams(height, forceRefresh)
	if err != nil {
		return pocket.Params{}, fmt.Errorf("ParamsAtHeight: provider error: %s", err)
	}
	var relaysToTokensMultiplier string
	var daoAlloc string
	var proposerCut string

	np := allParams.NodeParams
	relaysToTokensMultiplier, ok := np.Get("pos/RelaysToTokensMultiplier")

	if !ok {
		relaysToTokensMultiplier, _ = s.provider.Param("pos/RelaysToTokensMultiplier", height)

		if relaysToTokensMultiplier == "" {
			return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/RelaysToTokensMultiplier'", height)
		}
	}

	if params.RelaysToTokensMultiplier, ok = sdk.NewIntFromString(relaysToTokensMultiplier); !ok {
		return pocket.Params{}, errors.New("ParamsAtHeight: failed to parse node_params key 'pos/RelaysToTokensMultiplier'")
	}

	daoAlloc, ok = np.Get("pos/DAOAllocation")
	if !ok {
		daoAlloc, _ = s.provider.Param("pos/DAOAllocation", height)

		if daoAlloc == "" {
			return pocket.Params{}, errors.New("ParamsAtHeight: node_params key not found 'pos/DAOAllocation'")
		}
	}
	da, ok := sdk.NewIntFromString(daoAlloc)
	if !ok {
		return pocket.Params{}, errors.New("ParamsAtHeight: failed to parse node_params key 'pos/DAOAllocation")
	}
	params.DaoAllocation = da

	proposerCut, ok = np.Get("pos/ProposerPercentage")
	if !ok {
		proposerCut, _ = s.provider.Param("pos/ProposerPercentage", height)
		if proposerCut == "" {
			return pocket.Params{}, errors.New("ParamsAtHeight: node_params key not found 'pos/ProposerPercentage'")
		}
	}
	pp, ok := sdk.NewIntFromString(proposerCut)
	if !ok {
		return pocket.Params{}, errors.New("ParamsAtHeight: failed to parse node_params key 'pos/ProposerCut")
	}
	params.ProposerPercentage = pp

	claimExpirationBlocks, ok := allParams.PocketParams.Get("pocketcore/ClaimExpiration")
	if !ok {
		claimExpirationBlocks, _ = s.provider.Param("pocketcore/ClaimExpiration", height)
		if claimExpirationBlocks == "" {
			return pocket.Params{}, errors.New("ParamsAtHeight: node_params key not found 'pocketcore/ClaimExpiration'")
		}
	}
	claimExpires, ok := sdk.NewIntFromString(claimExpirationBlocks)
	if !ok {
		return pocket.Params{}, fmt.Errorf("ParamsAtHeight: failed to parse node_params ket 'pocketcore/ClaimExpiration': %s", err)
	}
	params.ClaimExpirationBlocks = claimExpires

	if height > pocket.RewardScalingActivationHeight {
		stakeWeightMultiplier, ok := np.Get("pos/ServicerStakeWeightMultiplier")
		if !ok {
			stakeWeightMultiplier, _ = s.provider.Param("pos/ServicerStakeWeightMultiplier", height)
			if stakeWeightMultiplier == "" {
				return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeWeightMultiplier'", height)
			}
		}
		if params.ServicerStakeWeightMultiplier, err = sdk.NewDecFromStr(stakeWeightMultiplier); err != nil {
			return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeWeightMultiplier", height)
		}

		stakeFloorMultiplier, ok := np.Get("pos/ServicerStakeFloorMultiplier")
		if !ok {
			stakeFloorMultiplier, _ = s.provider.Param("pos/ServicerStakeFloorMultiplier", height)
			if stakeFloorMultiplier == "" {
				return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeFloorMultiplier'", height)
			}
		}
		if params.ServicerStakeFloorMultiplier, ok = sdk.NewIntFromString(stakeFloorMultiplier); !ok {
			return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeFloorMultiplier", height)
		}

		stakeFloorMultiplierExponent, ok := np.Get("pos/ServicerStakeFloorMultiplierExponent")
		if !ok {
			stakeFloorMultiplierExponent, _ = s.provider.Param("pos/ServicerStakeFloorMultiplierExponent", height)
			if stakeFloorMultiplierExponent == "" {
				return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeFloorMultiplierExponent'", height)
			}
		}
		if params.ServicerStakeFloorMultiplierExponent, err = sdk.NewDecFromStr(stakeFloorMultiplierExponent); err != nil {
			return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeFloorMultiplierExponent", height)
		}

		stakeWeightCeiling, ok := np.Get("pos/ServicerStakeWeightCeiling")
		if !ok {
			stakeWeightCeiling, _ = s.provider.Param("pos/ServicerStakeWeightCeiling", height)
			if stakeWeightCeiling == "" {
				return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeWeightCeiling'", height)
			}
		}
		if params.ServicerStakeWeightCeiling, err = strconv.ParseInt(stakeWeightCeiling, 10, 64); err != nil {
			return pocket.Params{}, fmt.Errorf("ParamsAtHeight: node_params key not found at height %d 'pos/ServicerStakeWeightCeiling", height)
		}
	}

	return params, nil
}

func (s *Service) TxReward(tx pocket.Transaction) (pocket.Reward, error) {
	params, err := s.ParamsAtHeight(int64(tx.Height), false)
	if err != nil {
		return pocket.Reward{}, fmt.Errorf("TxReward: %s", err)
	}

	var stake sdk.BigInt
	var poktPerRelay, coinsDecimal, netCoinsDecimal, weight, percentageToKeep sdk.BigDec

	if tx.NumRelays.IsZero() {
		return pocket.Reward{}, nil
	}

	if tx.Height >= pocket.RewardScalingActivationHeight {
		node, err := s.provider.NodeAtHeight(tx.Address, tx.Height) // HERE
		if err != nil {
			return pocket.Reward{}, fmt.Errorf("TxReward: %s", err)
		}

		percentageToKeep = sdk.NewInt(100).
			Sub(params.ProposerPercentage).
			Sub(params.DaoAllocation).ToDec().Quo(sdk.NewDec(100))

		if params.ServicerStakeFloorMultiplier.GT(sdk.ZeroInt()) {
			const Pip22ExponentDenominator = 100

			stake = sdk.NewInt(int64(node.StakedBalance))

			flooredStake := sdk.MinInt(
				stake.Sub(stake.Mod(params.ServicerStakeFloorMultiplier)),
				sdk.NewInt(params.ServicerStakeWeightCeiling).
					Sub(sdk.NewInt(params.ServicerStakeWeightCeiling).
						Mod(params.ServicerStakeFloorMultiplier)),
			)
			// Convert from tokens to a BIN number
			bin := flooredStake.Quo(params.ServicerStakeFloorMultiplier)
			// calculate the weight value, weight will be a floatng point number so cast
			// to DEC here and then truncate back to big int
			weight = bin.ToDec().
				FracPow(
					params.ServicerStakeFloorMultiplierExponent,
					Pip22ExponentDenominator,
				).
				Quo(params.ServicerStakeWeightMultiplier)
			coinsDecimal = tx.NumRelays.Mul(params.RelaysToTokensMultiplier).ToDec().
				Mul(weight).Quo(sdk.NewDec(1000000))
			// truncate back to int

			poktPerRelay = coinsDecimal.Quo(tx.NumRelays.ToDec())
		}
	} else {
		weight = sdk.NewDec(1)
		poktPerRelay = params.LegacyPoktPerRelay()
		coinsDecimal = poktPerRelay.MulInt(tx.NumRelays)
	}

	netCoinsDecimal = coinsDecimal.Mul(percentageToKeep)

	return pocket.Reward{
		PoktAmount:    coinsDecimal,
		NetPoktAmount: netCoinsDecimal,
		StakeWeight:   weight,
		PoktPerRelay:  poktPerRelay,
	}, nil
}

func (s *Service) AccountTransactions(address string, page uint, perPage uint, sort string, transactionType string) ([]pocket.Transaction, error) {
	transactions := make([]pocket.Transaction, 0)

	pageIndex := page
	goAgain := true
	defaultPerPage := uint(1000)

	for goAgain {

		txs, err := s.provider.AccountTransactions(address, pageIndex, defaultPerPage, sort)

		if err != nil {
			return nil, fmt.Errorf("AccountTransactions: %s", err)
		}

		if len(txs) != int(defaultPerPage) {
			goAgain = false
			break
		}

		for _, tx := range txs {
			if transactionType == "" || tx.Type == transactionType {
				if len(transactions) == int(perPage) {
					goAgain = false
					break
				}

				params, err := s.ParamsAtHeight(int64(tx.Height), false)
				if err != nil {
					return nil, fmt.Errorf("AccountTransactions: %s", err)
				}

				tx.Time, err = s.provider.BlockTime(tx.Height)
				tx.Reward, err = s.TxReward(tx)
				if err != nil {
					return nil, fmt.Errorf("AccountTransactions: %s", err)
				}

				tx.PoktPerRelay = params.LegacyPoktPerRelay()

				tx.ExpireHeight = params.ClaimExpirationBlocks.Int64() + tx.Height

				transactions = append(transactions, tx)

			}

		}

		pageIndex++
	}

	return transactions, nil
}

func (s *Service) AccountClaimsAndProofs(address string) (claims, proofs map[string]pocket.Transaction, err error) {
	pageIndex := 1
	defaultPerPage := 10000
	sortDirection := "desc"
	keepFetching := true

	claims, proofs = make(map[string]pocket.Transaction), make(map[string]pocket.Transaction)

	for keepFetching {
		txs, err := s.provider.AccountTransactions(address, uint(pageIndex), uint(defaultPerPage), sortDirection)

		if len(txs) != defaultPerPage {
			keepFetching = false
		}

		if err != nil {
			continue
		}

		for _, tx := range txs {

			params, err := s.ParamsAtHeight(int64(tx.Height), false)
			tx.Time, err = s.provider.BlockTime(tx.Height)
			tx.Reward, err = s.TxReward(tx)

			if err != nil {
				continue
			}

			tx.PoktPerRelay = params.LegacyPoktPerRelay()

			tx.ExpireHeight = params.ClaimExpirationBlocks.Int64() + tx.Height

			sessionKey := sessionKey(tx)
			switch tx.Type {
			case pocket.TypeClaim:
				claims[sessionKey] = tx
				break
			case pocket.TypeProof:
				proofs[sessionKey] = tx
				break
			}

		}

		pageIndex++
	}

	return claims, proofs, nil
}

func (s *Service) Node(address string) (pocket.Node, error) {
	node, err := s.provider.Node(address)
	if err != nil {
		return pocket.Node{}, fmt.Errorf("Node: %s", err)
	}

	node.Balance, err = s.provider.Balance(address)
	if err != nil {
		return pocket.Node{}, fmt.Errorf("Node: %s", err)
	}

	//nodeProvider, err := s.provider.NodeProvider(node.Address)
	//if err != nil {
	//	log.Default().Printf("ERROR: %+v", err)
	//} else {
	//	if node.LatestBlockHeight, err = nodeProvider.Height(); err != nil {
	//		log.Default().Printf("ERROR: %+v", err)
	//	} else {
	//		blockTimes, err := s.BlockTimes([]uint{node.LatestBlockHeight})
	//		if err != nil {
	//			log.Default().Printf("ERROR: %+v", err)
	//		} else {
	//			node.LatestBlockTime = blockTimes[node.LatestBlockHeight]
	//		}
	//	}
	//}

	return node, nil
}

func (s *Service) SimulateRelay(servicerUrl, chainID string, payload map[string]interface{}) (json.RawMessage, error) {
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("SimulateRelay: %s", err)
	}

	resp, err := s.provider.SimulateRelay(servicerUrl, chainID, encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("SimulateRelay: %s", err)
	}

	return resp, nil
}

func (s *Service) RewardsByMonth(address string) (map[string]pocket.MonthlyReward, error) {
	claims, proofs, err := s.AccountClaimsAndProofs(address)
	if err != nil {
		return nil, fmt.Errorf("RewardsByMonth: %s", err)
	}

	months := make(map[string]pocket.MonthlyReward)
	for sessionKey, tx := range claims {
		tx.IsConfirmed = false
		proof, proofExists := proofs[sessionKey]
		if proofExists && proof.ResultCode == 0 {
			tx.IsConfirmed = true
		}

		monthKey := fmt.Sprintf("%d-%d", tx.Time.Year(), tx.Time.Month())
		if _, exists := months[monthKey]; !exists {
			months[monthKey] = pocket.MonthlyReward{
				Year:        uint(tx.Time.Year()),
				Month:       uint(tx.Time.Month()),
				TotalProofs: sdk.ZeroInt(),
				DaysOfWeek:  make(map[int]*pocket.DayOfWeek, 7),
			}
			months[monthKey].DaysOfWeek[0] = &pocket.DayOfWeek{Name: "Sunday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[1] = &pocket.DayOfWeek{Name: "Monday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[2] = &pocket.DayOfWeek{Name: "Tuesday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[3] = &pocket.DayOfWeek{Name: "Wednesday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[4] = &pocket.DayOfWeek{Name: "Thursday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[5] = &pocket.DayOfWeek{Name: "Friday", Proofs: sdk.ZeroInt()}
			months[monthKey].DaysOfWeek[6] = &pocket.DayOfWeek{Name: "Saturday", Proofs: sdk.ZeroInt()}
		}
		month := months[monthKey]
		if tx.IsConfirmed {
			month.TotalProofs = month.TotalProofs.Add(tx.NumRelays)
		}

		month.Transactions = append(month.Transactions, tx)
		months[monthKey] = month
	}

	for monthKey, mo := range months {
		sort.Slice(months[monthKey].Transactions, func(i, j int) bool {
			return mo.Transactions[i].Time.Before(mo.Transactions[j].Time)
		})

		var numTxs = float64(0)
		var totalSecs = float64(0)
		var prevTx, emptyYx = pocket.Transaction{}, pocket.Transaction{}
		for _, tx := range months[monthKey].Transactions {
			if prevTx != emptyYx {
				totalSecs += tx.Time.Sub(prevTx.Time).Seconds()
				numTxs++
			}
			prevTx = tx

			dayOfWeek := int(tx.Time.Weekday())
			months[monthKey].DaysOfWeek[dayOfWeek].Proofs = months[monthKey].DaysOfWeek[dayOfWeek].Proofs.Add(tx.NumRelays)
		}
		mo.AvgSecsBetweenRewards = totalSecs / numTxs
		if math.IsNaN(mo.AvgSecsBetweenRewards) {
			mo.AvgSecsBetweenRewards = 0
		}

		mo.TotalSecsBetweenRewards = totalSecs
		months[monthKey] = mo

	}

	return months, nil
}

func sessionKey(tx pocket.Transaction) string {
	return fmt.Sprintf("%d%s%s", tx.SessionHeight, tx.AppPubkey, tx.ChainID)
}
