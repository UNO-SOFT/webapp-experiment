// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"html/template"
	"net/http"

	"github.com/rs/xmux"

	"golang.org/x/net/context"
)

type User struct {
	Name string
}

var usersTmpl *template.Template

// usersGET lists the users
func usersGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	usersTmpl.ExecuteTemplate(w, "users.html", nil)
}

// userGET returns the user
func userGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userid := xmux.Param(ctx, "userid")
	usersTmpl.ExecuteTemplate(w, "user.html", map[string]string{
		"userid": userid,
	})
}

// userPOST creates a new user
func userPOST(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userid := r.PostFormValue("user")
	message := r.PostFormValue("message")
	//user(userid).Submit(userid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": userid + ": " + message,
	})
}

// userPUT modifies an existing user
func userPUT(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userid := r.PostFormValue("user")
	message := r.PostFormValue("message")
	//user(userid).Submit(userid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": userid + ": " + message,
	})
}

// userDELETE deletes the user
func userDELETE(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//userid := c.Param("userid")
	//deleteBroadcast(userid)
}
