package main

import (
	"log"

	"erp/internal/bootstrap"
)

// @title ERP Management API
// @version 1.0
// @description 电脑、打印机、办公设备、监控工程、网络工程、维修服务行业 ERP 后台 API。
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	app, err := bootstrap.NewApp("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
