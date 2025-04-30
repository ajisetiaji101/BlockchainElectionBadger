package repository

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func GetLastHash(db *badger.DB) ([]byte, error) {
	var lastHash []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return lastHash, nil
}

// get lh jika tidak ada buat lh baru
func GetLastHashOrCreate(db *badger.DB) error {

	return db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("lh"))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// Key tidak ditemukan, buat key baru
				return txn.Set([]byte("lh"), []byte("0"))
			}
			return err
		}
		return nil // Key sudah ada, tidak perlu melakukan apa-apa
	})
}

func SetLastHash(db *badger.DB, hash []byte) error {
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("lh"), hash)
	})
	if err != nil {
		return err
	}
	return nil
}

func SetDataRiwayat(db *badger.DB, key []byte, value []byte) error {
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	if err != nil {
		return err
	}
	return nil
}

func GetDataRiwayat(db *badger.DB, key []byte) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		fmt.Println("Key:", key)
		fmt.Println("Item:", item)

		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = val
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}
