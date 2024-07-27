package pocket

import (
	"fmt"
	sdk "monitoring-service/types"
	"strconv"

	"monitoring-service/pocket"
)

type accountTransactionsRequest struct {
	Address string `json:"address"`
	Height  uint   `json:"height"`
	Page    uint   `json:"page"`
	PerPage uint   `json:"per_page"`
	Sort    string `json:"order"`
}

type accountTransactionsResponse struct {
	Transactions []transactionResponse `json:"txs"`
}

type transactionRequest struct {
	Hash string `json:"hash"`
}

type txResult struct {
	Code int64 `json:"code"`
}

type transactionResponse struct {
	Hash   string        `json:"hash"`
	Height int64         `json:"height"`
	StdTx  stdTxResponse `json:"stdTx"`
	Result txResult      `json:"tx_result"`
}

func (t *transactionResponse) Transaction() (pocket.Transaction, error) {
	numProofs := sdk.NewInt(0)
	var err error

	if t.StdTx.Message.Value.TotalProofs != "" {
		parsedNum, ok := sdk.NewIntFromString(t.StdTx.Message.Value.TotalProofs)
		if !ok {
			return pocket.Transaction{}, fmt.Errorf("transactionResponse.Transaction: %s", err)
		}
		numProofs = parsedNum
	}

	tx := pocket.Transaction{
		Hash:       t.Hash,
		Height:     t.Height,
		Type:       t.StdTx.Message.Type,
		ChainID:    t.StdTx.Message.Value.Header.Chain,
		NumRelays:  numProofs,
		ResultCode: t.Result.Code,
	}

	switch tx.Type {
	case pocket.TypeProof:
		sessionHeight, err := strconv.ParseUint(t.StdTx.Message.Value.Leaf.Value.SessionHeight, 10, 32)
		if err != nil {
			return pocket.Transaction{}, fmt.Errorf("transactionResponse.Transaction: %s", err)
		}
		tx.SessionHeight = uint(sessionHeight)
		tx.AppPubkey = t.StdTx.Message.Value.Leaf.Value.AAT.AppPubkey
		tx.ChainID = t.StdTx.Message.Value.Leaf.Value.Blockchain
		break

	case pocket.TypeClaim:
		sessionHeight, err := strconv.ParseUint(t.StdTx.Message.Value.Header.SessionHeight, 10, 32)
		if err != nil {
			return pocket.Transaction{}, fmt.Errorf("transactionResponse.Transaction: %s", err)
		}

		tx.SessionHeight = uint(sessionHeight)
		tx.ChainID = t.StdTx.Message.Value.Header.Chain
		tx.AppPubkey = t.StdTx.Message.Value.Header.AppPubKey
		break
	}

	return tx, nil
}

type stdTxResponse struct {
	Fee     []feeResponse `json:"fee"`
	Message msgResponse   `json:"msg"`
}

type feeResponse struct {
	Amount       string `json:"amount"`
	Denomination string `json:"denom"`
}

type msgResponse struct {
	Type  string           `json:"type"`
	Value msgValueResponse `json:"value"`
}

type msgValueResponse struct {
	FromAddress string            `json:"from_address"`
	Header      msgHeaderResponse `json:"header"`
	TotalProofs string            `json:"total_proofs"`
	Leaf        msgLeafResponse   `json:"leaf"`
}

type msgLeafResponse struct {
	Type  string               `json:"type"`
	Value msgLeafValueResponse `json:"value"`
}

type msgLeafValueResponse struct {
	Blockchain    string             `json:"blockchain"`
	SessionHeight string             `json:"session_block_height"`
	AAT           msgLeafAATResponse `json:"aat"`
}

type msgLeafAATResponse struct {
	AppPubkey string `json:"app_pub_key"`
}

type msgHeaderResponse struct {
	AppPubKey     string `json:"app_public_key"`
	Chain         string `json:"chain"`
	SessionHeight string `json:"session_height"`
}
