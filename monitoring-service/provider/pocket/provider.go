package pocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	pchttp "monitoring-service/http"
	"monitoring-service/pocket"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
)

const (
	contentTypeJSON               = "application/json; charset=UTF-8"
	urlPathGetAccountTransactions = "query/accounttxs"
	urlPathGetTransaction         = "query/tx"
	urlPathGetBlock               = "query/block"
	urlPathGetNode                = "query/node"
	urlPathGetBalance             = "query/balance"
	urlPathGetHeight              = "query/height"
	urlPathGetParam               = "query/param"
	urlPathGetAllParams           = "query/allParams"
	urlPathSimulateRelay          = "v1/client/sim"
)

type blockTimesRepo interface {
	Get(height int64) (t time.Time, exists bool, err error)
	Set(height int64, t time.Time) error
}

type paramsRepo interface {
	Get(name string, height int64) (p pocket.Params, exists bool, err error)
	Set(name string, height int64, p pocket.Params) error
	GetAll(height int64) (params pocket.AllParams, exists bool, err error)
	SetAll(height int64, params pocket.AllParams) error
	DelAll(height int64) error
}

type nodesRepo interface {
	Get(addr string, height int64) (n pocket.Node, exists bool, err error)
	Set(addr string, height int64, node pocket.Node) error
}

type Provider interface {
	NodeProvider(address string) (Provider, error)
	Height() (uint, error)
	Param(name string, height int64) (string, error)
	AllParams(height int64, forceRefresh bool) (pocket.AllParams, error)
	Node(address string) (pocket.Node, error)
	NodeAtHeight(address string, height int64) (pocket.Node, error)
	Balance(address string) (uint, error)
	BlockTime(height int64) (time.Time, error)
	Transaction(hash string) (pocket.Transaction, error)
	AccountTransactions(address string, page uint, perPage uint, sort string) ([]pocket.Transaction, error)
	SimulateRelay(servicerUrl, chainID string, payload json.RawMessage) (json.RawMessage, error)
	WithLogger(l log.Logger) Provider
}

type pocketProvider struct {
	client         pchttp.Client
	blockTimesRepo blockTimesRepo
	paramsRepo     paramsRepo
	nodesRepo      nodesRepo
	pocketRpcURL   string
}

func NewPocketProvider(c pchttp.Client, pocketRpcURL string, blockTimesRepo blockTimesRepo, paramsRepo paramsRepo, nodesRepo nodesRepo) Provider {

	return pocketProvider{
		client:         c,
		blockTimesRepo: blockTimesRepo,
		paramsRepo:     paramsRepo,
		nodesRepo:      nodesRepo,
		pocketRpcURL:   pocketRpcURL,
	}
}

func (p pocketProvider) WithLogger(l log.Logger) Provider {
	return loggingProvider{
		provider: p,
		logger:   l,
	}
}

func (p pocketProvider) NodeProvider(addr string) (Provider, error) {
	node, err := p.Node(addr)
	if err != nil {
		return pocketProvider{}, err
	}

	return NewPocketProvider(p.client, fmt.Sprintf("%s/v1", node.ServiceURL), p.blockTimesRepo, p.paramsRepo, p.nodesRepo), nil
}

func (p pocketProvider) Height() (uint, error) {
	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetHeight)
	//var req interface{}
	var resp struct {
		Height float64 `json:"height"`
	}

	body, err := p.doRequest(url, nil)
	if err != nil {
		return 0, fmt.Errorf("pocketProvider.Height: %s", err)
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, fmt.Errorf("pocketProvider.Height: %s", err)
	}

	return uint(resp.Height), nil
}

// Param returns the value of a given parameter at the specified height. A height of 0 means the latest block.
func (p pocketProvider) Param(name string, height int64) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("pocketProvider.Param(%s): %s", name, err)
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetParam)
	pReq := paramRequest{
		Key:    name,
		Height: height,
	}
	var pRes paramResponse
	body, err := p.doRequest(url, pReq)
	if err != nil {
		fail(err)
	}

	if err := json.Unmarshal(body, &pRes); err != nil {
		fail(err)
	}

	return pRes.Value, nil
}

type allParamsRequest struct {
	Height int64 `json:"height"`
}

// AllParams returns all network parameters.
func (p pocketProvider) AllParams(height int64, forceRefresh bool) (pocket.AllParams, error) {
	fail := func(err error) (pocket.AllParams, error) {
		return pocket.AllParams{}, fmt.Errorf("pocketProvider.AllParams(%d): %s", height, err)
	}

	if forceRefresh {
		if err := p.paramsRepo.DelAll(height); err != nil {
			return pocket.AllParams{}, err
		}
	}

	cached, exists, err := p.paramsRepo.GetAll(height)
	if err == nil && exists {
		return cached, nil
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetAllParams)
	pReq := allParamsRequest{
		Height: height,
	}
	var pRes pocket.AllParams
	body, err := p.doRequest(url, pReq)
	if err != nil {
		return fail(err)
	}

	if err := json.Unmarshal(body, &pRes); err != nil {
		return fail(fmt.Errorf("unmarshal allParamsResponse: %s", err))
	}

	if err := p.paramsRepo.SetAll(height, pRes); err != nil {
		return fail(err)
	}
	return pRes, nil
}

func (p pocketProvider) NodeAtHeight(address string, height int64) (pocket.Node, error) {
	var fail = func(err error) (pocket.Node, error) {
		return pocket.Node{}, fmt.Errorf("pocketProvider.Node: %s", err)
	}

	node, exists, err := p.nodesRepo.Get(address, height)
	if exists && err == nil {
		return node, nil
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetNode)
	nodeRequest := queryNodeRequest{
		Address: address,
		Height:  height,
	}
	var nodeResponse queryNodeResponse

	body, err := p.doRequest(url, nodeRequest)
	if err != nil {
		return fail(err)
	}

	if err = json.Unmarshal(body, &nodeResponse); err != nil {
		return fail(err)
	}

	chains := make([]pocket.Chain, len(nodeResponse.Chains))
	for i, chainID := range nodeResponse.Chains {
		ch, err := pocket.ChainFromID(chainID)
		if err != nil {
			fail(err)
		}

		chains[i] = ch
	}

	stakedBal, err := strconv.ParseUint(nodeResponse.StakedBalance, 10, 64)
	if err != nil {
		return pocket.Node{}, fmt.Errorf("pocketProvider.Node: %s", err)
	}

	n := pocket.Node{
		Address:       nodeResponse.Address,
		Pubkey:        nodeResponse.Pubkey,
		ServiceURL:    nodeResponse.ServiceURL,
		StakedBalance: uint(stakedBal),
		IsJailed:      nodeResponse.IsJailed,
		Chains:        chains,
		IsSynced:      false,
	}

	err = p.nodesRepo.Set(address, height, n)
	if err != nil {
		return pocket.Node{}, err
	}
	return n, nil
}

func (p pocketProvider) Node(address string) (pocket.Node, error) {
	var fail = func(err error) (pocket.Node, error) {
		return pocket.Node{}, fmt.Errorf("pocketProvider.Node: %s", err)
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetNode)
	nodeRequest := queryNodeRequest{Address: address}
	var nodeResponse queryNodeResponse

	body, err := p.doRequest(url, nodeRequest)
	if err != nil {
		return fail(err)
	}

	if err = json.Unmarshal(body, &nodeResponse); err != nil {
		return fail(err)
	}

	chains := make([]pocket.Chain, len(nodeResponse.Chains))
	for i, chainID := range nodeResponse.Chains {
		ch, err := pocket.ChainFromID(chainID)
		if err != nil {
			fail(err)
		}

		chains[i] = ch
	}

	stakedBal, err := strconv.ParseUint(nodeResponse.StakedBalance, 10, 64)
	if err != nil {
		return pocket.Node{}, fmt.Errorf("pocketProvider.Node: %s", err)
	}

	return pocket.Node{
		Address:       nodeResponse.Address,
		Pubkey:        nodeResponse.Pubkey,
		ServiceURL:    nodeResponse.ServiceURL,
		StakedBalance: uint(stakedBal),
		IsJailed:      nodeResponse.IsJailed,
		Chains:        chains,
		IsSynced:      false,
		CustodialNode: nodeResponse.OutputAddress == address || nodeResponse.OutputAddress == "",
		OutputAddress: nodeResponse.OutputAddress,
	}, nil
}

func (p pocketProvider) Balance(address string) (uint, error) {
	var fail = func(err error) (uint, error) {
		return 0, fmt.Errorf("pocketProvider.Balance: %s", err)
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetBalance)
	balRequest := balanceRequest{Address: address}
	var balResponse balanceResponse

	body, err := p.doRequest(url, balRequest)
	if err != nil {
		return fail(err)
	}

	err = json.Unmarshal(body, &balResponse)
	if err != nil {
		return fail(err)
	}

	return balResponse.Balance, nil
}

func (p pocketProvider) BlockTime(height int64) (time.Time, error) {
	var fail = func(err error) (time.Time, error) {
		return time.Time{}, fmt.Errorf("pocketProvider.BlockTime: %s", err)
	}

	cached, exists, _ := p.blockTimesRepo.Get(height)
	if exists {
		return cached, nil
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetBlock)
	blkRequest := blockRequest{Height: height}
	var blkResponse blockResponse

	body, err := p.doRequest(url, blkRequest)
	if err != nil {
		return fail(err)
	}

	err = json.Unmarshal(body, &blkResponse)
	if err != nil {
		return fail(err)
	}

	if err = p.blockTimesRepo.Set(height, blkResponse.Block.Header.Time); err != nil {
		return time.Time{}, fmt.Errorf("pocketProvider.BlockTime: %s", err)
	}

	return blkResponse.Block.Header.Time, nil
}

func (p pocketProvider) Transaction(hash string) (pocket.Transaction, error) {
	var fail = func(err error) (pocket.Transaction, error) {
		return pocket.Transaction{}, fmt.Errorf("pocketProvider.Transaction: %s", err)
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetTransaction)
	txRequest := transactionRequest{Hash: hash}
	var txnResponse transactionResponse

	body, err := p.doRequest(url, txRequest)
	if err != nil {
		return fail(err)
	}

	err = json.Unmarshal(body, &txnResponse)
	if err != nil {
		return fail(err)
	}

	txn, err := txnResponse.Transaction()
	if err != nil {
		return fail(err)
	}

	return txn, nil
}

func (p pocketProvider) AccountTransactions(address string, page uint, perPage uint, sort string) ([]pocket.Transaction, error) {
	var fail = func(err error) ([]pocket.Transaction, error) {
		return nil, fmt.Errorf("pocketProvider.AccountTransactions: %s", err)
	}

	url := fmt.Sprintf("%s/%s", p.pocketRpcURL, urlPathGetAccountTransactions)
	txsRequest := accountTransactionsRequest{
		Address: address,
		Height:  0,
		Page:    page,
		PerPage: perPage,
		Sort:    sort,
	}
	var txsResponse accountTransactionsResponse

	body, err := p.doRequest(url, txsRequest)
	if err != nil {
		return fail(err)
	}

	err = json.Unmarshal(body, &txsResponse)
	if err != nil {
		return fail(err)
	}

	var transactions []pocket.Transaction
	for _, t := range txsResponse.Transactions {
		txn, err := t.Transaction()
		txn.Address = address
		if err != nil {
			return fail(err)
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
}

func (p pocketProvider) SimulateRelay(servicerUrl, chainID string, payload json.RawMessage) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/%s", servicerUrl, urlPathSimulateRelay)
	path := ""

	switch chainID {
	case "0003":
		path = "/ext/info"
	case "0001":
		path = "/v1/query/height"
	}

	simRequest := relayRequest{
		RelayNetworkID: chainID,
		Payload: relayRequestPayload{
			Data:    string(payload),
			Method:  "POST",
			Path:    path,
			Headers: make(map[string]string, 0),
		},
	}

	resp, err := p.doRequest(url, simRequest)
	if err != nil {
		return nil, fmt.Errorf("pocketProvider.SimulateRelay: %s", err)
	}

	return resp, nil
}

func (p pocketProvider) doRequest(url string, reqObj interface{}) ([]byte, error) {
	var err error
	var reqBody = make([]byte, 0)

	if reqObj != nil {
		reqBody, err = json.Marshal(reqObj)
		if err != nil {
			return nil, fmt.Errorf("doRequest: %s", err)
		}
	}

	req := bytes.NewBuffer(reqBody)
	clientReq, err := http.NewRequest(http.MethodPost, url, req)
	if err != nil {
		return nil, fmt.Errorf("doRequest: %s", err)
	}

	clientReq.Header.Set("Content-Type", contentTypeJSON)

	resp, err := p.client.Do(clientReq)
	if err != nil {
		return nil, fmt.Errorf("doRequest: %s", err)
	}

	if resp == nil {
		return nil, errors.New("pocketProvider.doRequest: got empty response for " + url)
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("pocketProvider.doRequest: got unexpected response status %s - %s", resp.Status, string(reqBody)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("doRequest: %s", err)
	}

	return body, nil
}
