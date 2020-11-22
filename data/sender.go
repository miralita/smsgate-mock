package data

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type Sender struct {
	SenderUuid uuid.UUID `json:"senderUuid"`
	Login string `json:"login"`
	Password string `json:"password,omitempty"`
}

func (s *Sender) Bytes() []byte {
	bindata, _ := json.Marshal(s)
	return bindata
}

func (s *Sender) FromBytes(bindata []byte) error {
	if err := json.Unmarshal(bindata, s); err != nil {
		return err
	}
	return nil
}

func getSenderBuckets(tx *bbolt.Tx) (bucketSendersLogins *bbolt.Bucket, bucketSenders *bbolt.Bucket, err error) {
	err = nil
	if bucketSendersLogins = tx.Bucket([]byte(BucketSendersByLogin)); bucketSendersLogins == nil {
		err = fmt.Errorf("can't load bucket %s", BucketSendersByLogin)
	}
	if bucketSenders = tx.Bucket([]byte(BucketSenders)); bucketSenders == nil {
		err = fmt.Errorf("Can't load bucket %s", BucketSenders)
	}
	return
}

func (s *Sender) Save(db *bbolt.DB) error {
	s.SenderUuid = uuid.New()
	err := db.Update(func(tx *bbolt.Tx) error {
		bucketSendersLogins, bucketSenders, err := getSenderBuckets(tx)
		if err != nil {
			return nil
		}
		existing := bucketSendersLogins.Get([]byte(s.Login))
		if existing != nil {
			return fmt.Errorf("duplicate login %s", s.Login)
		}
		if err := bucketSenders.Put(s.SenderUuid[:], s.Bytes()); err != nil {
			return fmt.Errorf("can't create sender: %v", err)
		}
		if err := bucketSendersLogins.Put([]byte(s.Login), s.SenderUuid[:]); err != nil {
			return fmt.Errorf("can't create login index: %v", err)
		}
		return nil
	})
	return err
}

func (s *Sender) Delete(db *bbolt.DB, id uuid.UUID) error {
	err := db.Update(func(tx *bbolt.Tx) error {
		bucketSendersLogins, bucketSenders, err := getSenderBuckets(tx)
		if err != nil {
			return nil
		}
		existing := bucketSenders.Get(id[:])
		if existing == nil {
			return fmt.Errorf("sender not found")
		}
		if err = s.FromBytes(existing); err != nil {
			return fmt.Errorf("can't parse existing sender data: %v", err)
		}
		if err = bucketSendersLogins.Delete([]byte(s.Login)); err != nil {
			return fmt.Errorf("can't delete sender login: %v", err)
		}
		if err = bucketSenders.Delete(id[:]); err != nil {
			return fmt.Errorf("can't delete sender data: %v", err)
		}
		return nil
	})
	return err
}

func (s *Sender) Edit(db *bbolt.DB) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucketSendersLogins, bucketSenders, err := getSenderBuckets(tx)
		if err != nil {
			return nil
		}
		bindata := bucketSenders.Get(s.SenderUuid[:])
		if bindata == nil {
			return fmt.Errorf("sender not found")
		}
		existing := &Sender{}
		if err = existing.FromBytes(bindata); err != nil {
			return fmt.Errorf("can't parse existing data: %v, %s", err, string(bindata))
		}
		if len(s.Login) > 0 && existing.Login != s.Login {
			if err = bucketSendersLogins.Delete([]byte(existing.Login)); err != nil {
				return fmt.Errorf("can't delete old index %s: %v", existing.Login, err)
			}
			if err = bucketSendersLogins.Put([]byte(s.Login), s.SenderUuid[:]); err != nil {
				return fmt.Errorf("can't save new index %s: %v", s.Login, err)
			}
			existing.Login = s.Login
		}
		if len(s.Password) > 0 && existing.Password !=s.Password {
			existing.Password = s.Password
		}
		if err = bucketSenders.Put(s.SenderUuid[:], existing.Bytes()); err != nil {
			return fmt.Errorf("can't save sender: %v", err)
		}
		return nil
	})
}

func (s *Sender) List(db *bbolt.DB) ([]*Sender, error) {
	var res []*Sender
	err := db.View(func(tx *bbolt.Tx) error {
		buckerSenders := tx.Bucket([]byte(BucketSenders))
		if buckerSenders == nil {
			return fmt.Errorf("can't load bucket %s", BucketSenders)
		}
		iterator := buckerSenders.Cursor()

		for k, v := iterator.First(); k != nil; k, v = iterator.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			sender := &Sender{}
			if err := sender.FromBytes(v); err != nil {
				return fmt.Errorf("can't parse sender: %v, %s", err, string(v))
			}
			res = append(res, sender)
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func (s *Sender)LoadById(db *bbolt.DB, id uuid.UUID) error {
	return db.View(func(tx *bbolt.Tx) error {
		buckerSenders := tx.Bucket([]byte(BucketSenders))
		if buckerSenders == nil {
			return fmt.Errorf("can't load bucket %s", BucketSenders)
		}
		bindata := buckerSenders.Get(id[:])
		if bindata == nil {
			return fmt.Errorf("sender not found")
		}
		if err := s.FromBytes(bindata); err != nil {
			return fmt.Errorf("can't parse sender data: %v, %s", err, string(bindata))
		}
		return nil
	})
}

func (s *Sender)LoadByLogin(db *bbolt.DB, login string) error {
	return db.View(func(tx *bbolt.Tx) error {
		bucketSendersLogins, bucketSenders, err := getSenderBuckets(tx)
		if err != nil {
			return nil
		}
		bindata := bucketSendersLogins.Get([]byte(login))
		if bindata == nil {
			return fmt.Errorf("sender not found: %s", login)
		}
		id, err := uuid.FromBytes(bindata)
		if err != nil {
			return fmt.Errorf("can't parse sender id from db: %v", err)
		}
		bindata = bucketSenders.Get(id[:])
		if bindata == nil {
			return fmt.Errorf("sender not found by id: %s", id.String())
		}
		if err := s.FromBytes(bindata); err != nil {
			return fmt.Errorf("can't parse sender data: %v, %s", err, string(bindata))
		}
		return nil
	})
}
