// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import "github.com/gin-gonic/gin"

// fundersGET lists the funders
func fundersGET(c *gin.Context) {
	c.HTML(200, "funders.html", nil)
}

// funderGET returns the funder
func funderGET(c *gin.Context) {
	funderid := c.Param("funderid")
	c.HTML(200, "funder.html", gin.H{
		"funderid": funderid,
	})
}

// funderPOST creates a new funder
func funderPOST(c *gin.Context) {
	funderid := c.PostForm("funder")
	message := c.PostForm("message")
	//funder(funderid).Submit(funderid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": funderid + ": " + message,
	})
}

// funderPUT modifies an existing funder
func funderPUT(c *gin.Context) {
	funderid := c.PostForm("funder")
	message := c.PostForm("message")
	//funder(funderid).Submit(funderid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": funderid + ": " + message,
	})
}

// funderDELETE deletes the funder
func funderDELETE(c *gin.Context) {
	//funderid := c.Param("funderid")
}
