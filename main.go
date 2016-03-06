// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"github.com/rs/xmux"
	"gopkg.in/authboss.v0"
	aboauth "gopkg.in/authboss.v0/oauth2"
)

var (
	ab       = authboss.New()
	database = NewMemStorer()
)

func main() {
	setupAuthboss()
	// Set up a middleware handler for Gin, with a custom "permission denied" message.
	mux := xmux.New()
	c := xhandler.Chain{}
	logConf := xlog.Config{}
	c.UseC(xlog.NewHandler(logConf))
	log.SetFlags(0)
	log.SetOutput(xlog.New(logConf))
	c.UseC(xhandler.CloseHandler)
	c.UseC(xhandler.TimeoutHandler(10 * time.Second))
	c.UseC(xlog.RemoteAddrHandler("ip"))
	c.UseC(xlog.UserAgentHandler("user_agent"))
	c.UseC(xlog.URLHandler("url"))
	c.UseC(xlog.RefererHandler("referer"))
	c.UseC(xlog.RequestIDHandler("req_id", "Request-Id"))

	h := func(f func(context.Context, http.ResponseWriter, *http.Request)) xhandler.HandlerC {
		return c.HandlerCF(xhandler.HandlerFuncC(f))
	}
	mux.GET("/", h(rootGET))

	admin := mux.NewGroup("/admin")
	admin.GET("/users", h(usersGET))
	sub := admin.NewGroup("/user")
	sub.GET("/:userid", h(userGET))
	sub.POST("/:userid", h(userPOST))
	sub.DELETE("/:userid", h(userDELETE))

	admin.GET("/raisers", h(raisersGET))
	sub = admin.NewGroup("/raiser")
	sub.GET("/:raiserid", h(raiserGET))
	sub.POST("/:raiserid", h(raiserPOST))
	sub.DELETE("/:raiserid", h(raiserDELETE))

	admin.GET("/funders", h(fundersGET))
	sub = admin.NewGroup("/funder")
	sub.GET("/:funderid", h(funderGET))
	sub.POST("/:funderid", h(funderPOST))
	sub.DELETE("/:funderid", h(funderDELETE))

	log.Fatal(http.ListenAndServe(":8080", c.Handler(mux)))
}

func rootGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var msg string
	uI, err := ab.CurrentUser(w, r)
	if err != nil {
		msg = fmt.Sprintf("CurrentUser: %v", err)
	} else {
		msg = fmt.Sprintf("%#v", uI)
	}
	io.WriteString(w, msg)
}

//go:generate go run gen_keys.go
func setupAuthboss() {
	cookieStore = securecookie.New(cookieStoreKey, nil)
	sessionStore = sessions.NewCookieStore(sessionStoreKey)

	ab.LogWriter = os.Stderr
	ab.Storer = database
	ab.CookieStoreMaker = NewCookieStorer
	ab.SessionStoreMaker = NewSessionStorer
	ab.OAuth2Storer = database
	ab.MountPath = "/auth"
	ab.OAuth2Providers = map[string]authboss.OAuth2Provider{
		"google": authboss.OAuth2Provider{
			OAuth2Config: &oauth2.Config{
				ClientID:     ``,
				ClientSecret: ``,
				Scopes:       []string{`profile`, `email`},
				Endpoint:     google.Endpoint,
			},
			Callback: aboauth.Google,
		},
	}
	ab.XSRFName = "csrf_token"
	ab.XSRFMaker = func(_ http.ResponseWriter, r *http.Request) string {
		return nosurf.Token(r)
	}
	ab.Policies = []authboss.Validator{
		authboss.Rules{
			FieldName:       "email",
			Required:        true,
			AllowWhitespace: false,
		},
		authboss.Rules{
			FieldName:       "password",
			Required:        true,
			MinLength:       6,
			MaxLength:       80,
			AllowWhitespace: true,
		},
	}

	if err := ab.Init(); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(ctx context.Context, w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		xlog.FromContext(ctx).Error(err)
	}
}
