package api

import (
	"github.com/google/uuid"
	"smsgate-mock/data"
	"time"
)

type MessageIn struct {
	Login string `json:"login"`
	Password string `json:"password"`
	SenderName string `json:"senderName"`
	MessageType string `json:"messageType"`
	MessageText string `json:"messageText"`
	ExpirationTimeout int `json:"expirationTimeout"`
	PhoneNumber string `json:"phoneNumber"`
}

func (s *MessageIn) ToModel() *data.Message {
	return &data.Message{
		Sender: &data.Sender{
			Login: s.Login,
			Password: s.Password,
		},
		SenderName: s.SenderName,
		MessageType: s.MessageType,
		MessageText: s.MessageText,
		ExpirationTimeout: s.ExpirationTimeout,
		PhoneNumber: s.PhoneNumber,
	}
}

type MessageOut struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	Status string `json:"status"`
	Create time.Time `json:"created"`
}

func (s *MessageOut) FromModel(src *data.Message)  *MessageOut {
	s.MessageUuid = src.MessageUuid
	s.Status = src.Status
	s.Create = src.Create
	return s
}

type MessageStatusOut struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	Status string `json:"status"`
	Sent time.Time `json:"sent"`
}

func (s *MessageStatusOut) FromModel(src *data.Message) *MessageStatusOut {
	s.MessageUuid = src.MessageUuid
	s.Status = src.Status
	s.Sent = src.Sent
	return s
}

type ListMessageOut struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	MessageType string `json:"messageType"`
	MessageText string `json:"messageText"`
	ExpirationTimeout int `json:"expirationTimeout"`
	PhoneNumber string `json:"phoneNumber"`
	Sent time.Time `json:"sent"`
}

func (s *ListMessageOut) FromModel(src *data.Message) *ListMessageOut {
	s.MessageText = src.MessageText
	s.MessageUuid = src.MessageUuid
	s.MessageType = src.MessageType
	s.ExpirationTimeout = src.ExpirationTimeout
	s.PhoneNumber = src.PhoneNumber
	s.Sent = src.Sent
	return s
}
