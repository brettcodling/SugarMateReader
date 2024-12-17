package database

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/brettcodling/SugarMateReader/internal/directory"
)

var DB *bolt.DB

func init() {
	var err error
	DB, err = bolt.Open(directory.ConfigDir+"settings.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Settings"))
		if b == nil {
			var err error
			b, err = tx.CreateBucket([]byte("Settings"))
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
}

func Get(key string) string {
	var value string
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Settings"))
		value = string(b.Get([]byte(key)))
		return nil
	})
	return value
}

func Set(key, value string) error {
	return DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Settings"))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}
