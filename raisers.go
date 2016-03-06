// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"fmt"
	"math/rand"

	"github.com/gin-gonic/gin"
)

// raisersGET lists the fundraisers
func raisersGET(c *gin.Context) {
	c.HTML(200, "raisers.html", nil)
}

// raiserGET shows the fundraiser.
func raiserGET(c *gin.Context) {
	raiserid := c.Param("raiserid")
	userid := fmt.Sprint(rand.Int31())
	c.HTML(200, "raiser.html", gin.H{
		"raiserid": raiserid,
		"userid":   userid,
	})
}

// raiserPUT modifies an existing fundraiser
func raiserPUT(c *gin.Context) {
	raiserid := c.Param("raiserid")
	message := c.PostForm("message")
	//raiser(raiserid).Submit(userid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": raiserid + ": " + message,
	})
}

// raiserPOST creates a new fundraiser
func raiserPOST(c *gin.Context) {
	raiserid := c.Param("raiserid")
	message := c.PostForm("message")
	//raiser(raiserid).Submit(userid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": raiserid + ": " + message,
	})
}

// raiserDELETE deletes the fundraiser
func raiserDELETE(c *gin.Context) {
	//raiserid := c.Param("raiserid")
	//deleteBroadcast(raiserid)
}
