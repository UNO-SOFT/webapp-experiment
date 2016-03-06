// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import "github.com/gin-gonic/gin"

// usersGET lists the users
func usersGET(c *gin.Context) {
	c.HTML(200, "users.html", nil)
}

// userGET returns the user
func userGET(c *gin.Context) {
	userid := c.Param("userid")
	c.HTML(200, "user.html", gin.H{
		"userid": userid,
	})
}

// userPOST creates a new user
func userPOST(c *gin.Context) {
	userid := c.PostForm("user")
	message := c.PostForm("message")
	//user(userid).Submit(userid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": userid + ": " + message,
	})
}

// userPUT modifies an existing user
func userPUT(c *gin.Context) {
	userid := c.PostForm("user")
	message := c.PostForm("message")
	//user(userid).Submit(userid + ": " + message)

	c.JSON(200, gin.H{
		"status":  "success",
		"message": userid + ": " + message,
	})
}

// userDELETE deletes the user
func userDELETE(c *gin.Context) {
	//userid := c.Param("userid")
	//deleteBroadcast(userid)
}
