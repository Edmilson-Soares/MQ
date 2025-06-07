package mq

import (
	"fmt"
	"log"

	"go.etcd.io/bbolt"
)

type MQKV struct {
	name string
	db   *bbolt.DB
}

func NewMQKV(name string) *MQKV {
	db, err := bbolt.Open(name, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	kv := &MQKV{
		db:   db,
		name: name,
	}
	kv.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("store"))
		if err != nil {
			return err
		}
		return nil

	})
	return kv
}
func (kv *MQKV) BCreate(name string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
}
func (kv *MQKV) BDelete(name string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		if name == "store" {
			return fmt.Errorf("can't delete store")
		}
		err := tx.DeleteBucket([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
}

func (kv *MQKV) Set(key, value string) error {
	return kv.BSet("store", key, value)
}

func (kv *MQKV) Del(key string) error {
	return kv.BDel("store", key)
}
func (kv *MQKV) Get(key string) (string, error) {
	return kv.BGet("store", key)
}

////////////////

func (kv *MQKV) BSet(bucket, key, value string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		return bucket.Put([]byte(key), []byte(value))
	})
}

func (kv *MQKV) BDel(bucket, key string) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		return bucket.Delete([]byte(key))
	})
}
func (kv *MQKV) BGet(bucket, key string) (string, error) {
	value := []byte{}
	err := kv.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return fmt.Errorf("bucket não encontrado")
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	return string(value), err

}

func (kv *MQKV) BList(bucket string, filter func(k, v []byte) bool) (map[string]string, error) {
	result := make(map[string]string)

	err := kv.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
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
