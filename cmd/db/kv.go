package db

import (
	"fmt"

	"go.etcd.io/bbolt"
)

func (kv *NoSQL) BCreate(name string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("kv_" + name))
		if err != nil {
			return err
		}
		return nil
	})
}
func (kv *NoSQL) BDelete(name string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		if name == "store" {
			return fmt.Errorf("can't delete store")
		}
		err := tx.DeleteBucket([]byte("kv_" + name))
		if err != nil {
			return err
		}
		return nil
	})
}

func (kv *NoSQL) Set(key, value string) error {
	return kv.BSet("store", key, value)
}

func (kv *NoSQL) Del(key string) error {
	return kv.BDel("store", key)
}
func (kv *NoSQL) Get(key string) (string, error) {
	return kv.BGet("store", key)
}

////////////////

func (kv *NoSQL) BSet(bucket, key, value string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("kv_" + bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		return bucket.Put([]byte(key), []byte(value))
	})
}

func (kv *NoSQL) BDel(bucket, key string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("kv_" + bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		return bucket.Delete([]byte(key))
	})
}
func (kv *NoSQL) BGet(bucket, key string) (string, error) {
	value := []byte{}
	err := kv.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("kv_" + bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	return string(value), err

}

func (kv *NoSQL) BList(bucket string, filter func(k, v []byte) bool) (map[string]string, error) {
	result := make(map[string]string)

	err := kv.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("kv_" + bucket))
		if b == nil {
			return fmt.Errorf("bucket '%s' não encontrado", bucket)
		}

		return b.ForEach(func(k, v []byte) error {
			if filter == nil || filter(k, v) {
				result[string(k)] = string(v)
			}
			return nil
		})
	})

	return result, err
}
