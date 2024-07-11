package monitoring

import (
	"context"
	"fmt"
	"monitoring-service/pocket"
	sdk "monitoring-service/types"
	"sort"
	"time"

	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	Height              endpoint.Endpoint
	Params              endpoint.Endpoint
	Node                endpoint.Endpoint
	Transaction         endpoint.Endpoint
	AccountTransactions endpoint.Endpoint
	BlockTimes          endpoint.Endpoint
	MonthlyRewards      endpoint.Endpoint
}

type heightResponse struct {
	Height uint `json:"height"`
}

func HeightEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		height, err := svc.Height()
		if err != nil {
			return nil, fmt.Errorf("HeightEndpoint: %s", err)
		}

		response = heightResponse{Height: height}
		return response, nil
	}
}

type monthlyRewardsRequest struct {
	Address string `json:"address"`
}

type monthlyRewardsResponse struct {
	Year                   uint                       `json:"year"`
	Month                  uint                       `json:"month"`
	NumRelays              sdk.BigInt                 `json:"num_relays"`
	PoktAmount             sdk.BigDec                 `json:"pokt_amount"`
	RelaysByChain          []relaysByChain            `json:"relays_by_chain"`
	AvgSecBetweenRewards   float64                    `json:"avg_sec_between_rewards"`
	TotalSecBetweenRewards float64                    `json:"total_sec_between_rewards"`
	Transactions           []transactionResponse      `json:"transactions"`
	DaysOfWeek             map[int]daysOfWeekResponse `json:"days_of_week"`
}

type daysOfWeekResponse struct {
	Name      string     `json:"name"`
	NumProofs sdk.BigInt `json:"num_proofs"`
}

type relaysByChain struct {
	Chain     string     `json:"chain"`
	Name      string     `json:"name"`
	NumRelays sdk.BigInt `json:"num_relays"`
}

func MonthlyRewardsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("MonthlyRewardsEndpoint: %s", err)
		}

		req, ok := request.(monthlyRewardsRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		months, err := svc.RewardsByMonth(req.Address)
		if err != nil {
			return fail(err)
		}

		resp := make([]monthlyRewardsResponse, len(months))
		i := 0
		for _, month := range months {
			resp[i] = monthlyRewardsResponse{
				Year:                   month.Year,
				Month:                  month.Month,
				NumRelays:              month.TotalProofs,
				PoktAmount:             month.PoktAmount(),
				RelaysByChain:          make([]relaysByChain, 0),
				AvgSecBetweenRewards:   month.AvgSecsBetweenRewards,
				TotalSecBetweenRewards: month.TotalSecsBetweenRewards,
				Transactions:           make([]transactionResponse, len(month.Transactions)),
			}

			byChain := make(map[string]sdk.BigInt)
			for j, tx := range month.Transactions {
				if _, isSet := byChain[tx.ChainID]; !isSet {
					byChain[tx.ChainID] = sdk.NewInt(0)
				}

				byChain[tx.ChainID] = byChain[tx.ChainID].Add(tx.NumRelays)
				chain, _ := tx.Chain()
				resp[i].Transactions[j] = transactionResponse{
					Hash:    tx.Hash,
					Height:  tx.Height,
					Time:    tx.Time,
					Type:    tx.Type,
					ChainID: tx.ChainID,
					Chain: chainResponse{
						Name: chain.Name,
						ID:   chain.ID,
					},
					SessionHeight: tx.SessionHeight,
					ExpireHeight:  tx.ExpireHeight,
					AppPubkey:     tx.AppPubkey,
					NumRelays:     tx.NumRelays,
					Reward: rewardResponse{
						Amount:       tx.Reward.PoktAmount,
						NetAmount:    tx.Reward.NetPoktAmount,
						StakeWeight:  tx.Reward.StakeWeight,
						PoktPerRelay: tx.Reward.PoktPerRelay,
					},
					IsConfirmed: tx.IsConfirmed,
				}
			}

			for ch, num := range byChain {
				byChainResp := relaysByChain{
					Chain:     ch,
					NumRelays: num,
				}

				chain, err := pocket.ChainFromID(ch)
				if err != nil {
					byChainResp.Name = ch
				} else {
					byChainResp.Name = chain.Name
				}

				resp[i].RelaysByChain = append(resp[i].RelaysByChain, byChainResp)
			}

			resp[i].DaysOfWeek = make(map[int]daysOfWeekResponse, len(month.DaysOfWeek))
			for j, d := range month.DaysOfWeek {
				resp[i].DaysOfWeek[j] = daysOfWeekResponse{
					Name:      d.Name,
					NumProofs: d.Proofs,
				}
			}
			i++
		}

		sort.Slice(resp, func(i, j int) bool {
			if resp[i].Year == resp[j].Year {
				return resp[i].Month > resp[j].Month
			}
			return resp[i].Year > resp[j].Year
		})

		return resp, nil
	}
}

type blockTimesRequest struct {
	Heights []int64 `json:"heights"`
}

type blockTimesResponse map[uint]time.Time

func BlockTimesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("BlockTimesEndpoint: %s", err)
		}

		req, ok := request.(blockTimesRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		blocks, err := svc.BlockTimes(req.Heights)
		if err != nil {
			return fail(err)
		}

		return blocks, nil
	}
}

type paramsRequest struct {
	Height       int64
	ForceRefresh bool
}

type paramsResponse struct {
	Value interface{}
}

func ParamsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("ParamsEndpoint: %s", err)
		}

		req, ok := request.(paramsRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		params, err := svc.ParamsAtHeight(req.Height, req.ForceRefresh)
		if err != nil {
			return fail(err)
		}

		return params, nil
	}
}

type transactionRequest struct {
	Hash string
}

type rewardResponse struct {
	Amount       sdk.BigDec `json:"amount"`
	NetAmount    sdk.BigDec `json:"net_amount"`
	StakeWeight  sdk.BigDec `json:"stake_weight"`
	PoktPerRelay sdk.BigDec `json:"pokt_per_relay"`
}

type transactionResponse struct {
	Hash          string         `json:"hash"`
	Height        int64          `json:"height"`
	Time          time.Time      `json:"time"`
	Type          string         `json:"type"`
	ChainID       string         `json:"chain_id"`
	Chain         chainResponse  `json:"chain"`
	SessionHeight uint           `json:"session_height"`
	ExpireHeight  int64          `json:"expire_height"`
	AppPubkey     string         `json:"app_pubkey"`
	NumRelays     sdk.BigInt     `json:"num_relays"`
	IsConfirmed   bool           `json:"is_confirmed"`
	Reward        rewardResponse `json:"reward"`
}

func TransactionEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("TransactionEndpoint: %s", err)
		}

		req, ok := request.(transactionRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		txn, err := svc.Transaction(req.Hash)

		if err != nil {
			return fail(err)
		}

		return transactionResponse{
			Hash:      txn.Hash,
			Height:    txn.Height,
			Time:      txn.Time,
			Type:      txn.Type,
			ChainID:   txn.ChainID,
			NumRelays: txn.NumRelays,
			Reward: rewardResponse{
				Amount:       txn.Reward.PoktAmount,
				NetAmount:    txn.Reward.NetPoktAmount,
				StakeWeight:  txn.Reward.StakeWeight,
				PoktPerRelay: txn.Reward.PoktPerRelay,
			},
		}, nil
	}
}

type accountTransactionsRequest struct {
	Address         string
	Page            uint
	PerPage         uint
	Sort            string
	TransactionType string
}

type accountTransactionsResponse []transactionResponse

func AccountTransactionsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("AccountTransactionsEndpoint: %s", err)
		}

		req, ok := request.(accountTransactionsRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		txs, err := svc.AccountTransactions(req.Address, req.Page, req.PerPage, req.Sort, req.TransactionType)
		if err != nil {
			return fail(err)
		}

		txsResponse := make(accountTransactionsResponse, len(txs))
		for i, tx := range txs {
			chain, _ := tx.Chain()
			txsResponse[i] = transactionResponse{
				Hash:    tx.Hash,
				Height:  tx.Height,
				Time:    tx.Time,
				Type:    tx.Type,
				ChainID: tx.ChainID,
				Chain: chainResponse{
					Name: chain.Name,
					ID:   chain.ID,
				},
				SessionHeight: tx.SessionHeight,
				ExpireHeight:  tx.ExpireHeight,
				AppPubkey:     tx.AppPubkey,
				NumRelays:     tx.NumRelays,
				IsConfirmed:   tx.ResultCode == 0,
				Reward: rewardResponse{
					Amount:       tx.Reward.PoktAmount,
					NetAmount:    tx.Reward.NetPoktAmount,
					StakeWeight:  tx.Reward.StakeWeight,
					PoktPerRelay: tx.Reward.PoktPerRelay,
				},
			}
		}

		return txsResponse, nil
	}
}

type relayRequest struct {
	ServicerURL string                 `json:"servicer_url"`
	ChainID     string                 `json:"chain_id"`
	Payload     map[string]interface{} `json:"payload"`
}

func (req relayRequest) validate() error {
	if req.ServicerURL == "" {
		return errors.Newf("relayRequest.validate: Missing required param 'servicer_url")
	}

	if req.ChainID == "" {
		return errors.Newf("relayRequest.validate: Missing required param 'chain_id")
	}

	if req.Payload == nil {
		return errors.Newf("relayRequest.validate: Missing required param 'payload")
	}

	return nil
}

func SimulateRelayEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("SimulateRelayEndpoint: %s", err)
		}

		req, ok := request.(relayRequest)
		if !ok {
			err := fmt.Errorf("SimulateRelayEndpoint failed to parse request: %v", request)
			return fail(err)
		}

		if err = req.validate(); err != nil {
			return fail(err)
		}

		res, err := svc.SimulateRelay(req.ServicerURL, req.ChainID, req.Payload)
		if err != nil {
			return fail(err)
		}

		return res, nil
	}
}

type nodeRequest struct {
	Address string `json:"address"`
}

type nodeResponse struct {
	Address           string          `json:"address"`
	Pubkey            string          `json:"pubkey"`
	ServiceURL        string          `json:"service_url"`
	Balance           uint            `json:"balance"`
	StakedBalance     uint            `json:"staked_balance"`
	IsJailed          bool            `json:"is_jailed"`
	Chains            []chainResponse `json:"chains"`
	IsSynced          bool            `json:"is_synced"`
	LatestBlockHeight uint            `json:"latest_block_height"`
	LatestBlockTime   time.Time       `json:"latest_block_time"`
	CustodialNode     bool            `json:"custodial_node"`
	OutputAddress     string          `json:"output_address"`
}

type chainResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func NodeEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fail := func(err error) (interface{}, error) {
			return nil, fmt.Errorf("NodeEndpoint: %s", err)
		}

		req, ok := request.(nodeRequest)
		if !ok {
			err := fmt.Errorf("failed to parse request: %v", request)
			return fail(err)
		}

		node, err := svc.Node(req.Address)
		if err != nil {
			return fail(err)
		}

		chains := make([]chainResponse, len(node.Chains))
		for i, c := range node.Chains {
			chains[i] = chainResponse{
				Name: c.Name,
				ID:   c.ID,
			}
		}

		return nodeResponse{
			Address:           node.Address,
			Balance:           node.Balance,
			Pubkey:            node.Pubkey,
			ServiceURL:        node.ServiceURL,
			StakedBalance:     node.StakedBalance,
			IsJailed:          node.IsJailed,
			Chains:            chains,
			IsSynced:          node.IsSynced,
			LatestBlockHeight: node.LatestBlockHeight,
			LatestBlockTime:   node.LatestBlockTime,
			CustodialNode:     node.CustodialNode,
			OutputAddress:     node.OutputAddress,
		}, nil
	}
}
