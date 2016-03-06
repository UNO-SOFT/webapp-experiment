// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"html/template"
	"net/http"

	"github.com/rs/xmux"

	"golang.org/x/net/context"
)

type Funder struct {
	User
}

var fundersTmpl *template.Template

// fundersGET lists the funders
func fundersGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	fundersTmpl.ExecuteTemplate(w, "funders.html", nil)
}

// funderGET returns the funder
func funderGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	funderid := xmux.Param(ctx, "funderid")
	fundersTmpl.ExecuteTemplate(w, "funder.html", map[string]string{
		"funderid": funderid,
	})
}

// funderPOST creates a new funder
func funderPOST(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	funderid := r.PostFormValue("funder")
	message := r.PostFormValue("message")
	//funder(funderid).Submit(funderid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": funderid + ": " + message,
	})
}

// funderPUT modifies an existing funder
func funderPUT(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	funderid := r.PostFormValue("funder")
	message := r.PostFormValue("message")
	//funder(funderid).Submit(funderid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": funderid + ": " + message,
	})
}

// funderDELETE deletes the funder
func funderDELETE(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//funderid := c.Param("funderid")
}
