// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"

	"golang.org/x/net/context"

	"github.com/rs/xmux"
)

type Raiser struct {
	User

	Want Money
}

var raisersTmpl *template.Template

type Money struct {
	Amount   int64
	Currency Currency
}

type Currency string

const (
	USD = Currency("USD")
	EUR = Currency("EUR")
	HUF = Currency("HUF")
)

// raisersGET lists the fundraisers
func raisersGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	raisersTmpl.ExecuteTemplate(w, "raisers.html", nil)
}

// raiserGET shows the fundraiser.
func raiserGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	raiserid := xmux.Param(ctx, "raiserid")
	userid := fmt.Sprint(rand.Int31())

	raisersTmpl.ExecuteTemplate(w, "raiser.html",
		map[string]string{
			"raiserid": raiserid,
			"userid":   userid,
		})
}

// raiserPUT modifies an existing fundraiser
func raiserPUT(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	raiserid := xmux.Param(ctx, "raiserid")
	message := r.PostFormValue("message")
	//raiser(raiserid).Submit(userid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": raiserid + ": " + message,
	})
}

// raiserPOST creates a new fundraiser
func raiserPOST(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	raiserid := xmux.Param(ctx, "raiserid")
	message := r.PostFormValue("message")
	//raiser(raiserid).Submit(userid + ": " + message)

	writeJSON(ctx, w, map[string]string{
		"status":  "success",
		"message": raiserid + ": " + message,
	})
}

// raiserDELETE deletes the fundraiser
func raiserDELETE(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//raiserid := c.Param("raiserid")
	//deleteBroadcast(raiserid)
}
