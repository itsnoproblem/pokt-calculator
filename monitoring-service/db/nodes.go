package db

import (
	"encoding/json"
	"fmt"

	"git.mills.io/prologic/bitcask"

	"monitoring-service/pocket"
)

type NodesRepo struct {
	db *bitcask.Bitcask
}

func NewNodesRepo(db *bitcask.Bitcask) NodesRepo {
	return NodesRepo{db: db}
}

func (r NodesRepo) Get(address string, height int64) (node pocket.Node, exists bool, err error) {
	keyB := r.key(address, height)
	nodeB, err := r.db.Get(keyB)
	if err != nil {
		return pocket.Node{}, false, fmt.Errorf("NodesRepo.Get [%s, %d]: %s", address, height, err)
	}

	if err = json.Unmarshal(nodeB, &node); err != nil {
		return pocket.Node{}, false, fmt.Errorf("NodesRepo.Get: failed to parse json for %s, %d: %s", address, height, err)
	}

	return node, true, nil
}

func (r NodesRepo) Set(address string, height int64, n pocket.Node) error {
	keyB := r.key(address, height)
	nodeB, _ := json.Marshal(n)
	err := r.db.Put(keyB, nodeB)
	if err != nil {
		return fmt.Errorf("NodesRepo.Set [%s, %d]: %s", address, height, err)
	}

	return nil
}

func (r NodesRepo) key(address string, height int64) []byte {
	key := fmt.Sprintf("%d%s", height, address)
	keyB, _ := json.Marshal(key)
	return keyB
}
