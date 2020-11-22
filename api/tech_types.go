package api

import (
	"github.com/google/uuid"
	"smsgate-mock/data"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

// input

type SenderIn struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

func (s *SenderIn) ToModel() *data.Sender {
	return &data.Sender{Login: s.Login,  Password: s.Password}
}

type SenderEditIn struct {
	Login string `json:"login"`
	Password string `json:"password"`
	SenderUuid uuid.UUID `json:"senderUuid"`
}

func (s *SenderEditIn) ToModel() *data.Sender {
	return &data.Sender{Login: s.Login, Password: s.Password, SenderUuid: s.SenderUuid}
}

// output

type SenderOut struct {
	SenderUuid uuid.UUID `json:"senderUuid"`
	Login string `json:"login"`
}

func (s *SenderOut) FromModel(src *data.Sender) *SenderOut {
	s.Login = src.Login
	s.SenderUuid = src.SenderUuid
	return s
}
