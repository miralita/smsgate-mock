package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
	"time"
)

type Message struct {
	MessageUuid       uuid.UUID
	Sender            *Sender `json:"-"`
	SenderUuid        uuid.UUID
	SenderName        string
	MessageType       string
	MessageText       string
	ExpirationTimeout int
	PhoneNumber       string
	Status            string
	Create            time.Time
	Sent              time.Time
}

func (s *Message) Bytes() []byte {
	bindata, _ := json.Marshal(s)
	return bindata
}

func (s *Message) FromBytes(bindata []byte) error {
	if err := json.Unmarshal(bindata, s); err != nil {
		return err
	}
	return nil
}

func (s *Message) Index() []byte {
	suf := []byte(s.Create.UTC().Format("2006-01-02 15:04:05.999"))
	// we need for reversed sort
	for i := 0; i < len(suf); i++ {
		suf[i] = 255 - suf[i]
	}
	return append([]byte(s.PhoneNumber), suf...)
}

func (s *Message) Save(db *bbolt.DB) error {
	s.MessageUuid = uuid.New()
	s.Create = time.Now()
	s.Sent = time.Now()
	s.Status = "SENT"
	s.SenderUuid = s.Sender.SenderUuid
	return db.Update(func(tx *bbolt.Tx) error {
		bucketMessages, bucketMessageIndex, err := s.GetMessageBuckets(tx)
		if err != nil {
			return err
		}
		if err := bucketMessages.Put(s.MessageUuid[:], s.Bytes()); err != nil {
			return fmt.Errorf("can't save message: %v", err)
		}
		idx := s.Index()
		if bindata := bucketMessageIndex.Get(idx); bindata != nil {
			return fmt.Errorf("message index already exists: possible throttling")
		}
		if err := bucketMessageIndex.Put(idx, s.MessageUuid[:]); err != nil {
			return fmt.Errorf("can't save message index: %v", err)
		}
		return nil
	})
}

func (s *Message) LoadById(db *bbolt.DB, id uuid.UUID) error {
	return db.View(func(tx *bbolt.Tx) error {
		bucketMessages := tx.Bucket([]byte(BucketMessages))
		if bucketMessages == nil {
			return fmt.Errorf("can't get bucket for messages")
		}
		bindata := bucketMessages.Get(id[:])
		if bindata == nil {
			return fmt.Errorf("message not found: %s", id.String())
		}
		if err := s.FromBytes(bindata); err != nil {
			return fmt.Errorf("can't parse message data: %v %s", err, string(bindata))
		}
		return nil
	})
}

func (s *Message) Delete(db *bbolt.DB, id uuid.UUID) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucketMessages, bucketMessageIndex, err := s.GetMessageBuckets(tx)
		if err != nil {
			return err
		}
		existing := bucketMessages.Get(id[:])
		if existing == nil {
			return fmt.Errorf("message not found")
		}
		if err := s.FromBytes(existing); err != nil {
			return fmt.Errorf("can't parse existing message data: %v %s", err, string(existing))
		}
		if err := bucketMessages.Delete(id[:]); err != nil {
			return fmt.Errorf("can't delete message: %v", err)
		}
		if err := bucketMessageIndex.Delete(s.Index()); err != nil {
			return fmt.Errorf("can't delete message index: %v", err)
		}
		return nil
	})
}

func (s *Message) GetMessageBuckets(tx *bbolt.Tx) (*bbolt.Bucket, *bbolt.Bucket, error) {
	bucketMessages := tx.Bucket([]byte(BucketMessages))
	if bucketMessages == nil {
		return nil, nil, fmt.Errorf("can't get bucket for messages")
	}
	bucketMessageIndex := tx.Bucket([]byte(BucketMessageIndex))
	if bucketMessageIndex == nil {
		return nil, nil, fmt.Errorf("can't get bucket for message index")
	}
	return bucketMessages, bucketMessageIndex, nil
}

func (s *Message) ListByPhone(db *bbolt.DB, phone string) ([]*Message, error) {
	ret := make([]*Message, 0)
	err := db.View(func(tx *bbolt.Tx) error {
		bucketMessages, bucketMessageIndex, err := s.GetMessageBuckets(tx)
		if err != nil {
			return err
		}
		iterator := bucketMessageIndex.Cursor()
		prefix := []byte(phone)
		for k, v := iterator.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = iterator.Next() {
			msg, err := s.GetMessageFromBucket(bucketMessages, v, k)
			if err != nil {
				return err
			}
			ret = append(ret, msg)
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}

func (s *Message) GetMessageFromBucket(bucketMessages *bbolt.Bucket, messageUuid []byte, messageIndex []byte) (*Message, error) {
	bindata := bucketMessages.Get(messageUuid)
	if bindata == nil {
		return nil, fmt.Errorf("can't find message by id %v from index %v", messageUuid, messageIndex)
	}
	msg := &Message{}
	if err := msg.FromBytes(bindata); err != nil {
		return nil, fmt.Errorf("can't parse message: %v, %s", err, string(bindata))
	}
	return msg, nil
}

func (s *Message) List(db *bbolt.DB, limit, offset int) ([]*Message, error) {
	ret := make([]*Message, 0)
	err := db.View(func(tx *bbolt.Tx) error {
		bucketMessages, bucketMessageIndex, err := s.GetMessageBuckets(tx)
		if err != nil {
			return err
		}
		iterator := bucketMessageIndex.Cursor()
		if limit == 0 {
			limit = 10
		}
		limit += offset
		i := 0
		for k, v := iterator.First(); k != nil && i < limit; k, v = iterator.Next() {
			i += 1
			if i <= offset {
				continue
			}
			msg, err := s.GetMessageFromBucket(bucketMessages, v, k)
			if err != nil {
				return err
			}
			ret = append(ret, msg)
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}
