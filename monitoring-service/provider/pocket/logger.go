package pocket

import (
	"encoding/json"
	"fmt"
	"monitoring-service/pocket"
	"monitoring-service/timer"
	"time"

	log "github.com/go-kit/kit/log"
)

const (
	logTypeError = "ERROR"
	logTypeWarn  = "WARN"
	logTypeInfo  = "INFO"
)

type ProviderWithLogger interface {
	Provider
	WithLogger(l log.Logger) Provider
}

type loggingProvider struct {
	provider Provider
	logger   log.Logger
}

func (p loggingProvider) WithLogger(l log.Logger) Provider {
	//l = log.With(l, "ts", log.DefaultTimestampUTC)
	return loggingProvider{
		provider: p,
		logger:   l,
	}
}

func (p loggingProvider) NodeProvider(addr string) (Provider, error) {
	return p.provider.NodeProvider(addr)
}

func (p loggingProvider) Height() (uint, error) {
	t := timer.Start()
	h, err := p.provider.Height()
	if err != nil {
		p.error(err.Error())
	}

	p.info("Height is %d (took %s)", h, t.Elapsed().String())
	return h, nil
}

func (p loggingProvider) AllParams(height int64, forceRefresh bool) (pocket.AllParams, error) {
	t := timer.Start()
	params, err := p.provider.AllParams(height, forceRefresh)
	if err != nil {
		p.error(err.Error())
	}

	p.info("AllParams at height %d (took %s)", height, t.Elapsed().String())
	return params, nil
}

func (p loggingProvider) Param(name string, height int64) (string, error) {
	t := timer.Start()
	param, err := p.provider.Param(name, height)
	if err != nil {
		p.error(err.Error())
	}

	p.info("Param %s at height %d is %s (took %s)", name, height, param, t.Elapsed().String())
	return param, nil
}

func (p loggingProvider) Node(address string) (pocket.Node, error) {
	t := timer.Start()
	n, err := p.provider.Node(address)
	if err != nil {
		p.error(err.Error())
		return pocket.Node{}, err
	}

	p.info("Node for address %s (took %s)", address, t.Elapsed().String())
	return n, nil
}

func (p loggingProvider) NodeAtHeight(address string, height int64) (pocket.Node, error) {
	t := timer.Start()
	n, err := p.provider.NodeAtHeight(address, height)
	if err != nil {
		p.error(err.Error())
		return pocket.Node{}, err
	}

	p.info("NodeAtHeight for address %s (took %s)", address, t.Elapsed().String())
	return n, nil
}

func (p loggingProvider) Balance(address string) (uint, error) {
	t := timer.Start()
	b, err := p.provider.Balance(address)
	if err != nil {
		p.error(err.Error())
		return 0, err
	}

	p.info("Balance for address %s is %d (took %s)", address, b, t.Elapsed().String())
	return b, nil
}

func (p loggingProvider) BlockTime(height int64) (time.Time, error) {
	//t := timer.Start()
	bt, err := p.provider.BlockTime(height)
	if err != nil {
		p.error(err.Error())
		return time.Time{}, err
	}

	//p.info("BlockTime for %d (took %s)", height, t.Elapsed().String())
	return bt, nil
}

func (p loggingProvider) Transaction(hash string) (pocket.Transaction, error) {
	t := timer.Start()
	tx, err := p.provider.Transaction(hash)
	if err != nil {
		p.error(err.Error())
		return pocket.Transaction{}, nil
	}

	p.info("Transaction hash %s (took %s)", hash, t.Elapsed().String())
	return tx, nil
}

func (p loggingProvider) AccountTransactions(address string, page uint, perPage uint, sort string) ([]pocket.Transaction, error) {
	t := timer.Start()
	txs, err := p.provider.AccountTransactions(address, page, perPage, sort)
	if err != nil {
		p.error(err.Error())
		return nil, err
	}

	p.info("AccountTransactions for address %s page %d: %d results (took %s)",
		address, page, len(txs), t.Elapsed().String())
	return txs, nil
}

func (p loggingProvider) SimulateRelay(servicer_url, chainID string, payload json.RawMessage) (json.RawMessage, error) {
	t := timer.Start()
	res, err := p.provider.SimulateRelay(servicer_url, chainID, payload)
	p.info("SimulateRelay for %s: %s - %s (took %s)", chainID, servicer_url, string(payload), t.Elapsed())
	if err != nil {
		p.error(err.Error())
		return nil, err
	}

	return res, nil
}

func (p loggingProvider) error(format string, args ...interface{}) {
	p.log(logTypeError, format, args...)
}

func (p loggingProvider) warn(format string, args ...interface{}) {
	p.log(logTypeWarn, format, args...)
}

func (p loggingProvider) info(format string, args ...interface{}) {
	p.log(logTypeInfo, format, args...)
}

func (p loggingProvider) log(lvl, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_ = p.logger.Log("level", lvl, "msg", msg)
}
