package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"gopkg.in/go-playground/assert.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"smsgate-mock/api"
	"smsgate-mock/data"
	"smsgate-mock/utils"
	"testing"
)

func initApi(t *testing.T) *api.App {
	cfg := utils.ReadSettings()
	db, err := bbolt.Open(cfg.DbPath, 0600, nil)
	if err != nil {
		log.Fatalf("Can't open database %s: %v", cfg.DbPath, err)
	}
	data.InitBuckets(db)
	app := api.Init(cfg, db)
	return app
}

func TestMock(t *testing.T) {
	app := initApi(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/sender", bytes.NewBuffer([]byte(`{"login": "test", "password":"123"}`)))
	app.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
	res := api.SenderOut{}
	json.Unmarshal(w.Body.Bytes(), &res)

	w = httptest.NewRecorder()
	edit := &api.SenderEditIn{SenderUuid: res.SenderUuid, Password: "newpwd"}
	data, _ := json.Marshal(edit)
	req, _ = http.NewRequest("PATCH", "/api/v1/sender/" + res.SenderUuid.String(), bytes.NewBuffer(data))
	app.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)

	w = httptest.NewRecorder()
	edit = &api.SenderEditIn{SenderUuid: res.SenderUuid, Login: "new_login"}
	data, _ = json.Marshal(edit)
	req, _ = http.NewRequest("PATCH", "/api/v1/sender/" + res.SenderUuid.String(), bytes.NewBuffer(data))
	app.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/sender", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/sender/check_connection/" + res.SenderUuid.String(), bytes.NewBuffer([]byte(`{"login": "new_login", "password":"newpwd"}`)))
	app.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)

	msg := &api.MessageIn{
		Login:             "new_login",
		Password:          "newpwd",
		SenderName:        "TEST",
		MessageType:       "TEXT",
		MessageText:       "Test text ",
		ExpirationTimeout: 300,
		PhoneNumber:       "81234567890",
	}
	for i := 0; i < 18; i++ {
		msg.MessageText = fmt.Sprintf("%s %d", msg.MessageText, i)
		if i == 12 {
			msg.PhoneNumber = "81234567891"
		}
		data, _ = json.Marshal(msg)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/v1/message", bytes.NewBuffer(data))
		app.ServeHTTP(w, req)
		assert.Equal(t, 201, w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/message/search?phoneNumber=81234567891", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var msglist []*api.ListMessageOut
	err := json.Unmarshal(w.Body.Bytes(), &msglist)
	assert.Equal(t, nil, err)
	assert.Equal(t, 6, len(msglist))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/message/" + msglist[0].MessageUuid.String(), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/message?limit=10&offset=0", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	json.Unmarshal(w.Body.Bytes(), &msglist)
	assert.Equal(t, 10, len(msglist))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/message?limit=10&offset=10", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var msglist1 []*api.ListMessageOut
	json.Unmarshal(w.Body.Bytes(), &msglist1)
	assert.Equal(t, 8, len(msglist1))

	for i := 0; i < 10; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/api/v1/message/" + msglist[i].MessageUuid.String(), nil)
		app.ServeHTTP(w, req)
		assert.Equal(t, 204, w.Code)
	}

	for i := 0; i < 8; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/api/v1/message/" + msglist1[i].MessageUuid.String(), nil)
		app.ServeHTTP(w, req)
		assert.Equal(t, 204, w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/sender/" + res.SenderUuid.String(), nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)
}
