package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"smsgate-mock/data"
	"strconv"
	"strings"
)

// Message godoc
// @Summary Create new SMS
// @Param message body MessageIn true "Message data"
// @Success 201 {object} MessageOut
// @Failure 404 {object} ErrorMessage
// @Failure 422 {object} ErrorMessage
// @Failure 401 {object} ErrorMessage
// @Router /message [post]
func (app *App) Message(c *gin.Context) {
	var req MessageIn
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("can't parse JSON: %v", err))
		c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Can't parse request body"})
		return
	}
	sender := &data.Sender{}
	if err := sender.LoadByLogin(app.db, req.Login); err != nil {
		c.Error(fmt.Errorf("can't load sender: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find sender"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't send message due to internal server error"})
		}
		return
	}
	if sender.Password != req.Password {
		c.JSON(http.StatusUnauthorized, &ErrorMessage{"Password mismatch"})
		return
	}
	msg := req.ToModel()
	msg.Sender = sender
	if err := msg.Save(app.db); err != nil {
		c.Error(fmt.Errorf("can't save message: %v", err))
		c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't send message due to internal server error"})
		return
	}
	c.JSON(http.StatusCreated, (&MessageOut{}).FromModel(msg))
}

// MessageStatus godoc
// @Summary Get SMS status
// @Param messageUuid path string true "Message ID"
// @Success 200 {object} MessageStatusOut
// @Failure 404 {object} ErrorMessage
// @Failure 400 {object} ErrorMessage
// @Router /message/{messageUuid}/status [get]
func (app *App) MessageStatus(c *gin.Context) {
	idS := c.Param("messageUuid")
	if idS == "search" {
		// dirty hack due to gin-gonic routing model https://github.com/gin-gonic/gin/issues/1730
		app.SearchMessage(c)
		return
	}
	id, err := uuid.Parse(idS)
	if err != nil {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"Can't parse message uuid"})
		return
	}
	msg := &data.Message{}
	if err = msg.LoadById(app.db, id); err != nil {
		c.Error(fmt.Errorf("can'load message: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find message"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't get message due to internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, (&MessageStatusOut{}).FromModel(msg))
}

// DeleteMessage godoc
// @Summary Delete message
// @Param messageUuid path string true "Message ID"
// @Success 204
// @Failure 400 {object} ErrorMessage
// @Failure 404 {object} ErrorMessage
// @Failure 500 {object} ErrorMessage
// @Router /message/{messageUuid} [delete]
func (app *App) DeleteMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("messageUuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"Can't parse message uuid"})
		return
	}
	if err = (&data.Message{}).Delete(app.db, id); err != nil {
		c.Error(fmt.Errorf("can't delete message from database: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find requested message"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't delete message due to internal server error"})
		}
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

// SearchMessage godoc
// @Summary Search messages by phone number
// @Param phoneNumber query string true "Phone number"
// @Success 200 {array} ListMessageOut
// @Failure 500 {object} ErrorMessage
// @Router /message/search [get]
func (app *App) SearchMessage(c *gin.Context) {
	phoneNumber := c.Query("phoneNumber")
	retdata, err := (&data.Message{}).ListByPhone(app.db, phoneNumber)
	if err != nil {
		c.Error(fmt.Errorf("can't find messages by phone number: %v", err))
		c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't find messages due to internal server error"})
	}
	res := make([]*ListMessageOut, len(retdata))
	for i := 0; i < len(retdata); i++ {
		res[i] = (&ListMessageOut{}).FromModel(retdata[i])
	}
	c.JSON(http.StatusOK, res)
}

// ListMessage godoc
// @Summary List messages
// @Param limit query string false "Limit, default 10"
// @Param offset query string false "Offset, default 0"
// @Success 200 {array} ListMessageOut
// @Failure 422 {object} ErrorMessage
// @Failure 500 {object} ErrorMessage
// @Router /message [get]
func (app *App) ListMessage(c *gin.Context) {
	limitS := c.Query("limit")
	offsetS := c.Query("offset")
	limit := 10
	offset := 0
	var err error
	if len(limitS) > 0 {
		limit, err = strconv.Atoi(limitS)
		if err != nil {
			c.Error(fmt.Errorf("can't parse limit: %v", err))
			c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Bad limit"})
			return
		}
	}
	if len(offsetS) > 0 {
		offset, err = strconv.Atoi(offsetS)
		if err != nil {
			c.Error(fmt.Errorf("can't parse offset: %v", err))
			c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Bad offset"})
			return
		}
	}
	retdata, err := (&data.Message{}).List(app.db, limit, offset)
	if err != nil {
		c.Error(fmt.Errorf("can't list messages: %v", err))
		c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't list messages due to internal server error"})
	}
	res := make([]*ListMessageOut, len(retdata))
	for i := 0; i < len(retdata); i++ {
		res[i] = (&ListMessageOut{}).FromModel(retdata[i])
	}
	c.JSON(http.StatusOK, res)
}
