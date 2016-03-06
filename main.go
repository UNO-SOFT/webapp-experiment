// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

func main() {
	rootTmpl := template.Must(template.ParseGlob("templates/*.html"))
	root := gin.Default()
	root.SetHTMLTemplate(rootTmpl)
	root.GET("/", rootGET)

	adminTmpl := template.Must(template.ParseGlob("templates/admin/*.html"))
	admin := gin.Default()
	admin.SetHTMLTemplate(adminTmpl)
	admin.GET("/users", usersGET)
	admin.GET("/user/:userid", userGET)
	admin.POST("/user/:userid", userPOST)
	admin.DELETE("/user/:userid", userDELETE)

	admin.GET("/raisers", raisersGET)
	admin.GET("/raiser/:raiserid", raiserGET)
	admin.POST("/raiser/:raiserid", raiserPOST)
	admin.DELETE("/raiser/:raiserid", raiserDELETE)

	admin.GET("/funders", fundersGET)
	admin.GET("/funder/:funderid", funderGET)
	admin.POST("/funder/:funderid", funderPOST)
	admin.DELETE("/funder/:funderid", funderDELETE)

	root.Run(":8080")
}

func rootGET(c *gin.Context) {
	c.HTML(200, "root.html", nil)
}
