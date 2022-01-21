package monitoring

import (
	"context"
	"fmt"
	"monitoring-service/pocket"
	"sort"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	Height              endpoint.Endpoint
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
	Address string `json:"address"'`
}

type monthlyRewardsResponse struct {
	Year          uint                  `json:"year"`
	Month         uint                  `json:"month"`
	NumRelays     uint                  `json:"num_relays"`
	PoktAmount    float64               `json:"pokt_amount"`
	RelaysByChain []relaysByChain       `json:"relays_by_chain"`
	Transactions  []transactionResponse `json:"transactions"`
}

type relaysByChain struct {
	Chain     string `json:"chain"`
	Name      string `json:"name"`
	NumRelays uint   `json:"num_relays"`
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
				Year:          month.Year,
				Month:         month.Month,
				NumRelays:     month.TotalProofs,
				PoktAmount:    month.PoktAmount(),
				RelaysByChain: make([]relaysByChain, 0),
				Transactions:  make([]transactionResponse, len(month.Transactions)),
			}

			byChain := make(map[string]uint, 0)
			for j, tx := range month.Transactions {
				if _, isSet := byChain[tx.ChainID]; !isSet {
					byChain[tx.ChainID] = 0
				}

				byChain[tx.ChainID] += tx.NumProofs
				resp[i].Transactions[j] = transactionResponse{
					Hash:          tx.Hash,
					Height:        tx.Height,
					Time:          tx.Time,
					Type:          tx.Type,
					ChainID:       tx.ChainID,
					SessionHeight: tx.SessionHeight,
					AppPubkey:     tx.AppPubkey,
					NumProofs:     tx.NumProofs,
					IsConfirmed:   tx.IsConfirmed,
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
	Heights []uint `json:"heights"`
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

type transactionRequest struct {
	Hash string
}

type transactionResponse struct {
	Hash          string    `json:"hash"`
	Height        uint      `json:"height"`
	Time          time.Time `json:"time"`
	Type          string    `json:"type"`
	ChainID       string    `json:"chain_id"`
	SessionHeight uint      `json:"session_height"`
	AppPubkey     string    `json:"app_pubkey"`
	NumProofs     uint      `json:"num_proofs"`
	IsConfirmed   bool      `json:"is_confirmed"`
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
			NumProofs: txn.NumProofs,
		}, nil
	}
}

type accountTransactionsRequest struct {
	Address string
	Page    uint
	PerPage uint
	Sort    string
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

		txs, err := svc.AccountTransactions(req.Address, req.Page, req.PerPage, req.Sort)
		if err != nil {
			return fail(err)
		}

		txsResponse := make(accountTransactionsResponse, len(txs))
		for i, tx := range txs {
			txsResponse[i] = transactionResponse{
				Hash:          tx.Hash,
				Height:        tx.Height,
				Time:          tx.Time,
				Type:          tx.Type,
				ChainID:       tx.ChainID,
				SessionHeight: tx.SessionHeight,
				AppPubkey:     tx.AppPubkey,
				NumProofs:     tx.NumProofs,
			}
		}

		return txsResponse, nil
	}
}

type nodeRequest struct {
	Address string `json:"address"`
}

type nodeResponse struct {
	Address           string          `json:"address"`
	Pubkey            string          `json:"pubkey"`
	Balance           uint            `json:"balance"`
	StakedBalance     string          `json:"staked_balance"`
	IsJailed          bool            `json:"is_jailed"`
	Chains            []chainResponse `json:"chains"`
	IsSynced          bool            `json:"is_synced"`
	LatestBlockHeight uint            `json:"latest_block_height"`
	LatestBlockTime   time.Time       `json:"latest_block_time"`
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
			StakedBalance:     node.StakedBalance,
			IsJailed:          node.IsJailed,
			Chains:            chains,
			IsSynced:          node.IsSynced,
			LatestBlockHeight: node.LatestBlockHeight,
			LatestBlockTime:   node.LatestBlockTime,
		}, nil
	}
}
