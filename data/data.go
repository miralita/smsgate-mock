package data

import (
	"fmt"
	"go.etcd.io/bbolt"
	"log"
)

const (
	BucketSenders = "Senders"
	BucketSendersByLogin = "SendersByLogin"
	BucketMessages = "Messages"
	BucketMessageIndex = "MessageIndex"
)

func InitBuckets(db *bbolt.DB) {
	err := db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(BucketSenders)); err != nil {
			return fmt.Errorf("can't create bucket Senders: %v", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(BucketSendersByLogin)); err != nil {
			return fmt.Errorf("can't create index for Senders: %v", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(BucketMessages)); err != nil {
			return fmt.Errorf("can't create bucket Messages: %v", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(BucketMessageIndex)); err != nil {
			return fmt.Errorf("can't create bucket MessageIndex: %v", err)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Can't create buckets: %v", err)
	}
}
