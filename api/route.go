package api

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"smsgate-mock/utils"
	"strings"
	"time"
)

type App struct {
	cfg *utils.Settings
	r   *gin.Engine
	db  *bbolt.DB
}

func Init(cfg *utils.Settings, db *bbolt.DB) *App {
	app := &App{cfg: cfg, r: gin.New(), db: db}
	app.setupRoutes()
	return app
}

func (app *App) setupRoutes() {
	app.r.Use(gin.Recovery())
	if app.cfg.LogRequest {
		app.r.Use(RequestLoggerMiddleware())
	}
	if app.cfg.LogResponse {
		app.r.Use(ResponseLoggerMiddleware)
	}
	app.r.Use(gin.LoggerWithFormatter(customFormatter))
	api_r := app.r.Group("/api/v1")
	api_r.GET("/sender", app.ListSenders)
	api_r.POST("/sender", app.AddSender)
	api_r.DELETE("/sender/:senderUuid", app.DeleteSender)
	api_r.PATCH("/sender/:senderUuid", app.EditSender)
	api_r.POST("/sender/check_connection/:senderUuid", app.CheckConnection)
	api_r.POST("/message", app.Message)
	api_r.GET("/message", app.ListMessage)
	api_r.DELETE("/message/:messageUuid", app.DeleteMessage)
	api_r.GET("/message/:messageUuid", app.MessageStatus)
	app.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.r.ServeHTTP(w, r)
}

func (app *App) Run() {
	addr := fmt.Sprintf(":%d", app.cfg.ListenPort)
	app.r.Run(addr)
}

func customFormatter(param gin.LogFormatterParams) string {
	//                          1    2   3   4  5  6  7  8    9    10
	return fmt.Sprintf("[%s] [%s] %s %d %s %s %s %s \"%s\": %s\n",
		param.TimeStamp.Format(time.RFC3339),     // 1
		param.ClientIP,                           // 2
		param.Request.Header.Get("X-Request-Id"), // 3
		param.StatusCode,                         // 4
		param.Latency,                            // 5
		param.Method,                             // 6
		param.Request.URL,                        // 7
		param.Request.Proto,                      // 8
		param.Request.UserAgent(),                // 9
		param.ErrorMessage,                       // 10
	)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > 0 && strings.HasPrefix(c.Request.URL.String(), "/api/v1") {
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := ioutil.ReadAll(tee)
			c.Request.Body = ioutil.NopCloser(&buf)
			log.Println("Request body: " + string(body))
			//log.Println(c.Request.Header)
		}
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ResponseLoggerMiddleware(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	if blw.body.Len() > 0 && strings.HasPrefix(c.Request.URL.String(), "/api/v1") {
		log.Println("Response body: " + blw.body.String())
	}
}
