package pocket

import "time"

type blockRequest struct {
	Height uint `json:"height"`
}

type blockResponse struct {
	Block blockResponseBlock `json:"block"`
}

type blockResponseBlock struct {
	Hash   string              `json:"hash"`
	Header blockHeaderResponse `json:"header"`
}

type blockHeaderResponse struct {
	Proposer string    `json:"proposer_address"`
	Time     time.Time `json:"time"`
}
