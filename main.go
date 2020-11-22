package main

import (
	"go.etcd.io/bbolt"
	"log"
	"smsgate-mock/api"
	"smsgate-mock/data"
	_ "smsgate-mock/docs"
	"smsgate-mock/utils"
)

// @title SMS-gate Mock
// @version 1.0
// @description This is a simple emulator for SMS-gate
// @query.collection.format multi
// @BasePath /api/v1
// @x-extension-openapi {"example": "value on a json format"}
func main() {
	cfg := utils.ReadSettings()
	db, err := bbolt.Open(cfg.DbPath, 0600, nil)
	if err != nil {
		log.Fatalf("Can't open database %s: %v", cfg.DbPath, err)
	}
	data.InitBuckets(db)
	app := api.Init(cfg, db)
	app.Run()
}
