package monitoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"monitoring-service/pocket"
	"net/http"
	"strconv"

	"monitoring-service/api"

	"github.com/gorilla/mux"
)

const (
	heightEndpointPath              = "/height"
	paramsEndpointPath              = "/params/{height}"
	transactionEndpointPath         = "/transactions/{hash}"
	nodeEndpointPath                = "/node/{address}"
	accountTransactionsEndpointPath = "/accounts/{address}/transactions"
	blockTimesEndpointPath          = "/block-times"
	monthlyRewardsEndpointPath      = "/node/{address}/rewards"
	simulateRelaysEndpointPath      = "/tests/simulate-relay"
)

type transport struct {
	Service Service
	Routes  []api.Route
}

func NewTransport(svc Service) transport {
	return transport{
		Service: svc,
		Routes: []api.Route{
			{
				Method:   http.MethodGet,
				Path:     heightEndpointPath,
				Endpoint: HeightEndpoint(svc),
				Decoder:  api.DecodeEmptyRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Path:     paramsEndpointPath,
				Method:   http.MethodGet,
				Endpoint: ParamsEndpoint(svc),
				Decoder:  decodeParamsRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodGet,
				Path:     transactionEndpointPath,
				Endpoint: TransactionEndpoint(svc),
				Decoder:  decodeTransactionRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodGet,
				Path:     accountTransactionsEndpointPath,
				Endpoint: AccountTransactionsEndpoint(svc),
				Decoder:  decodeAccountTransactionsRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodGet,
				Path:     nodeEndpointPath,
				Endpoint: NodeEndpoint(svc),
				Decoder:  decodeNodeRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodPost,
				Path:     blockTimesEndpointPath,
				Endpoint: BlockTimesEndpoint(svc),
				Decoder:  decodeBlockTimesRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodGet,
				Path:     monthlyRewardsEndpointPath,
				Endpoint: MonthlyRewardsEndpoint(svc),
				Decoder:  decodeMonthlyRewardsRequest,
				Encoder:  api.EncodeResponse,
			},
			{
				Method:   http.MethodPost,
				Path:     simulateRelaysEndpointPath,
				Endpoint: SimulateRelayEndpoint(svc),
				Decoder:  decodeSimulateRelaysRequest,
				Encoder:  api.EncodeResponse,
			},
		},
	}
}

func decodeParamsRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	vars := mux.Vars(req)
	height, ok := vars["height"]
	if !ok {
		return nil, errors.New("decodeParamsRequest: required param 'height' not found")
	}
	h, err := strconv.ParseInt(height, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("decodeParamsRequest: failed to parse height: %s", err)
	}

	var forceRefresh bool
	force, ok := req.URL.Query()["refresh"]
	if ok && len(force[0]) > 0 {
		forceRefresh, err = strconv.ParseBool(force[0])
		if err != nil {
			return nil, fmt.Errorf("decodeParamsRequest: failed to parse refresh param: %s", err)
		}
	}

	return paramsRequest{
		Height:       h,
		ForceRefresh: forceRefresh,
	}, nil
}
func decodeTransactionRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	vars := mux.Vars(req)
	hash, ok := vars["hash"]
	if !ok {
		return nil, errors.New("decodeTransactionRequest: required param 'hash' not found")
	}

	return transactionRequest{Hash: hash}, nil
}

func decodeAccountTransactionsRequest(ctx context.Context, req *http.Request) (request interface{}, err error) {
	vars := mux.Vars(req)
	address, ok := vars["address"]
	if !ok {
		return nil, errors.New("decodeTransactionRequest: required param 'address' not found")
	}

	var page uint64
	reqPage := req.URL.Query().Get("page")
	if reqPage != "" {
		page, err = strconv.ParseUint(reqPage, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("decodeAccountTransactionsRequest: %s", err)
		}
	}

	var perPage uint64
	reqPerPage := req.URL.Query().Get("per_page")
	if reqPerPage != "" {
		perPage, err = strconv.ParseUint(reqPerPage, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("decodeAccountTransactionsRequest: %s", err)
		}
	}

	sort := req.URL.Query().Get("sort")
	if sort == "" {
		sort = "asc"
	}

	transactionType := req.URL.Query().Get("type")

	if transactionType == "claim" {
		transactionType = pocket.TypeClaim
	}

	if transactionType == "proof" {
		transactionType = pocket.TypeProof
	}

	return accountTransactionsRequest{
		Address:         address,
		Page:            uint(page),
		PerPage:         uint(perPage),
		Sort:            sort,
		TransactionType: transactionType,
	}, nil
}

func decodeNodeRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	vars := mux.Vars(req)
	address, ok := vars["address"]
	if !ok {
		return nil, errors.New("decodeNodeRequest: required param 'address' not found")
	}

	return nodeRequest{Address: address}, nil
}

func decodeBlockTimesRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	var blockHeights blockTimesRequest
	if err := json.NewDecoder(req.Body).Decode(&blockHeights); err != nil {
		return nil, fmt.Errorf("decodeBlockTimesRequest: %s", err)
	}

	return blockHeights, nil
}

func decodeMonthlyRewardsRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	vars := mux.Vars(req)
	address, ok := vars["address"]
	if !ok {
		return nil, errors.New("decodeMonthlyRewardsRequest: required param 'address' not found")
	}

	return monthlyRewardsRequest{Address: address}, nil
}

func decodeSimulateRelaysRequest(_ context.Context, req *http.Request) (request interface{}, err error) {
	var simRequest relayRequest
	if err := json.NewDecoder(req.Body).Decode(&simRequest); err != nil {
		return nil, fmt.Errorf("decodeSimulateRelayRequest: %s", err)
	}

	return simRequest, nil
}
