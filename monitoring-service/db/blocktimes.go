package db

import (
	"encoding/json"
	"fmt"
	"time"

	"git.mills.io/prologic/bitcask"
)

type BlockTimesRepo struct {
	db *bitcask.Bitcask
}

func NewBlockTimesRepo(db *bitcask.Bitcask) BlockTimesRepo {
	return BlockTimesRepo{db: db}
}

func (r BlockTimesRepo) Get(height int64) (t time.Time, exists bool, err error) {
	heightB, _ := json.Marshal(height)
	blkTimeB, err := r.db.Get(heightB)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("BlockTimesRepo.Get [%d]: %s", height, err)
	}

	var blockTime time.Time
	if err = json.Unmarshal(blkTimeB, &blockTime); err != nil {
		return time.Time{}, false, fmt.Errorf("BlockTimesRepo,Get: failed to parse json for %d: %s", height, err)
	}

	return blockTime, true, nil
}

func (r BlockTimesRepo) Set(height int64, t time.Time) error {
	heightB, _ := json.Marshal(height)
	timeB, _ := json.Marshal(t)
	err := r.db.Put(heightB, timeB)
	if err != nil {
		return fmt.Errorf("BlockTimesRepo.Set: %s", err)
	}

	return nil
}
