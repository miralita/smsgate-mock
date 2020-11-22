package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/deepcopier"
	"net/http"
	"smsgate-mock/data"
	"strings"
)

// AddSender godoc
// @Summary Create new sender
// @Produce json
// @Param sender body SenderIn true "New sender"
// @Success 201 {object} SenderOut
// @Failure 422 {object} ErrorMessage
// @Failure 409 {object} ErrorMessage
// @Failure 500 {object} ErrorMessage
// @Router /sender [post]
func (app *App) AddSender(c *gin.Context) {
	var req SenderIn
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("can't parse JSON: %v", err))
		c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Can't parse request body"})
		return
	}
	sender := req.ToModel()
	if err  := sender.Save(app.db); err != nil {
		c.Error(fmt.Errorf("can't save sender to database: %v", err))
		if strings.HasPrefix(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, &ErrorMessage{"Can't save sender: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't save sender due to internal server error"})
		}
		return
	}
	res := (&SenderOut{}).FromModel(sender)
	deepcopier.Copy(sender).To(res)
	c.JSON(http.StatusCreated, res)
}

// DeleteSender godoc
// @Summary Delete sender
// @Param senderUuid path string true "Sender ID"
// @Success 204
// @Failure 400 {object} ErrorMessage
// @Failure 404 {object} ErrorMessage
// @Failure 500 {object} ErrorMessage
// @Router /sender/{senderUuid} [delete]
func (app *App) DeleteSender(c *gin.Context) {
	id, err := uuid.Parse(c.Param("senderUuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"Can't parse sender uuid"})
		return
	}
	if err := (&data.Sender{}).Delete(app.db, id); err != nil {
		c.Error(fmt.Errorf("can't delete sender from database: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find requested sender"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't delete sender due to internal server error"})
		}
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

// EditSender godoc
// @Summary Edit sender
// @Param senderUuid path string true "Sender ID"
// @Param sender body SenderOut true "Edited sender"
// @Success 204
// @Failure 404 {object} ErrorMessage
// @Failure 400 {object} ErrorMessage
// @Failure 422 {object} ErrorMessage
// @Router /sender/{senderUuid} [patch]
func (app *App) EditSender(c *gin.Context) {
	id, err := uuid.Parse(c.Param("senderUuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"Can't parse sender uuid"})
		return
	}
	var req SenderEditIn
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("can't parse JSON: %v", err))
		c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Can't parse request body"})
		return
	}
	if id != req.SenderUuid {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"SenderUuid from URL != SenderUuid from data"})
		return
	}
	sender := &data.Sender{}
	deepcopier.Copy(req).To(sender)
	err = sender.Edit(app.db)
	if err != nil {
		c.Error(fmt.Errorf("can't edit sender: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find requested sender"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't edit sender due to internal server error"})
		}
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

// ListSenders godoc
// @Summary List senders
// @Success 200 {array} SenderOut
// @Failure 500 {object} ErrorMessage
// @Router /sender [get]
func (app *App)ListSenders(c *gin.Context) {
	retdata, err := (&data.Sender{}).List(app.db)
	if err != nil {
		c.Error(fmt.Errorf("can't list senders: %v", err))
		c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't list senders due to internal server error"})
		return
	}
	res := make([]*SenderOut, len(retdata))
	for i := 0; i < len(retdata); i++ {
		res[i] = (&SenderOut{}).FromModel(retdata[i])
	}
	c.JSON(http.StatusOK, res)
}

// CheckConnection godoc
// @Summary Check sender's login and password
// @Param senderUuid path string true "Sender ID"
// @Param sender body SenderIn true "Login and password"
// @Success 204
// @Failure 404 {object} ErrorMessage
// @Failure 401 {object} ErrorMessage
// @Failure 400 {object} ErrorMessage
// @Failure 422 {object} ErrorMessage
// @Router /sender/check_connection/{senderUuid} [post]
func (app *App)CheckConnection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("senderUuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &ErrorMessage{"Can't parse sender uuid"})
		return
	}
	var req SenderIn
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("can't parse JSON: %v", err))
		c.JSON(http.StatusUnprocessableEntity, &ErrorMessage{"Can't parse request body"})
		return
	}
	sender := &data.Sender{}
	if err = sender.LoadById(app.db, id); err != nil {
		c.Error(fmt.Errorf("can't load sender: %v", err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &ErrorMessage{"Can't find requested sender"})
		} else {
			c.JSON(http.StatusInternalServerError, &ErrorMessage{"Can't check sender due to internal server error"})
		}
		return
	}
	if req.Login != sender.Login {
		c.JSON(http.StatusUnauthorized, &ErrorMessage{"Login mismatch"})
	} else if req.Password != sender.Password {
		c.JSON(http.StatusUnauthorized, &ErrorMessage{"Password mismatch"})
	} else {
		c.JSON(http.StatusNoContent, gin.H{})
	}
}
