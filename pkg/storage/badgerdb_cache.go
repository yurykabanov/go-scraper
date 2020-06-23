package storage

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/yurykabanov/scraper/pkg/domain"
)

type BadgerDBCache struct {
	db *badger.DB
}

func NewBadgerDBCache(db *badger.DB) *BadgerDBCache {
	return &BadgerDBCache{
		db: db,
	}
}

func (c *BadgerDBCache) Get(hash string) (*domain.Result, error) {
	var data []byte

	err := c.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(hash))
		if err != nil {
			return err
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && err != badger.ErrKeyNotFound {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	data, err = unpack(data)
	if err != nil {
		return nil, err
	}

	var result domain.Result
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *BadgerDBCache) Put(hash string, resp *domain.Result) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	data, err = pack(data)
	if err != nil {
		return err
	}

	return c.db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(hash), data)
	})
}
