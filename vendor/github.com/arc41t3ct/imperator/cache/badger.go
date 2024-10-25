package cache

import (
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

func (c *BadgerCache) Has(cacheKey string) (bool, error) {
	if _, err := c.Get(cacheKey); err != nil {
		return false, nil
	}
	return true, nil
}

func (c *BadgerCache) Get(cacheKey string) (interface{}, error) {
	var fromCache []byte

	err := c.Conn.View(
		func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(cacheKey))
			if err != nil {
				return err
			}

			err = item.Value(
				func(val []byte) error {
					fromCache = append([]byte{}, val...)
					return nil
				},
			)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}
	item := decoded[cacheKey]
	return item, nil
}

func (c *BadgerCache) Set(cacheKey string, value interface{}, expires ...int) error {
	entry := Entry{}
	entry[cacheKey] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}
	if len(expires) > 0 {
		err = c.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(cacheKey), encoded).WithTTL(time.Second * time.Duration(expires[0]))
			err = txn.SetEntry(e)
			return err
		})
	} else {
		err = c.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(cacheKey), encoded)
			err = txn.SetEntry(e)
			return err
		})
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *BadgerCache) Forget(cacheKey string) error {
	err := c.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(cacheKey))
		return err
	})
	return err
}

func (c *BadgerCache) EmptyMatching(cacheKey string) error {
	return c.emptyByMatch(cacheKey)
}

func (c *BadgerCache) Empty() error {
	return c.emptyByMatch("")
}

func (c *BadgerCache) emptyByMatch(cacheKey string) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := c.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000
	err := c.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		iter := txn.NewIterator(opts)
		defer iter.Close()
		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for iter.Seek([]byte(cacheKey)); iter.ValidForPrefix([]byte(cacheKey)); iter.Next() {
			key := iter.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}
