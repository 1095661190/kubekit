package controllers

import (
	"kubekit/utils"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

type MainRouter struct {
	router *gin.Engine
}

func StartToolkitServer() {
	r := gin.Default()

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	mainRouter := &MainRouter{}
	mainRouter.Initialize(r)
}

func (self *MainRouter) Initialize(r *gin.Engine) {

	self.router = r
	self.router.GET("/", self.IndexHandler)
	self.router.GET("/node/list", self.ListNodesHandler)

	color.Green("\r\n%sToolkit server is listening at: 0.0.0.0:9000", utils.CheckSymbol)
	self.router.Run(":9000")
}
