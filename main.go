// Copyright 2016 Tamás Gulácsi. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/authboss.v0"
	_ "gopkg.in/authboss.v0/auth"
	_ "gopkg.in/authboss.v0/confirm"
	_ "gopkg.in/authboss.v0/lock"
	aboauth "gopkg.in/authboss.v0/oauth2"
	_ "gopkg.in/authboss.v0/recover"
	_ "gopkg.in/authboss.v0/register"

	"github.com/UNO-SOFT/webapp-experiment/tpl"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rjeczalik/notify"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"github.com/rs/xmux"
)

var (
	ab       = authboss.New()
	database = NewMemStorer()

	funcs = template.FuncMap{
		"formatDate": func(date time.Time) string {
			return date.Format("2006/01/02 03:04pm")
		},
		"yield": func() string { return "" },
	}
	templates = tpl.Must(tpl.Load("views", "views/partials", "layout.html.tpl", funcs))
)

func main() {
	logConf := xlog.Config{
		Level:  xlog.LevelDebug,
		Output: xlog.NewConsoleOutput(),
	}
	logger := xlog.New(logConf)
	log.SetFlags(0)
	log.SetOutput(logger)

	// Set up a middleware handler for Gin, with a custom "permission denied" message.
	setupAuthboss(logger)

	fsEvents := make(chan notify.EventInfo, 1)
	for _, path := range []string{"templates"} {
		if err := notify.Watch(path+"/...", fsEvents, //recursive
			notify.InCloseWrite, notify.InMovedTo, notify.InCreate, notify.InDelete,
		); err != nil {
			logger.Errorf("cannot watch %q: %v", path, err)
		}
	}
	defer notify.Stop(fsEvents)
	go func() {
		var timer *time.Timer
		var timerC <-chan time.Time
		for {
			select {
			case event := <-fsEvents:
				if timerC == nil {
					logger.Warnf("%q changed (%s)", event.Path(), event.Event())
					timer = time.NewTimer(2 * time.Second)
					timerC = timer.C
					continue
				}
				timer.Reset(2 * time.Second)
			case <-timerC:
				logger.Warn("RESTARTING")
				syscall.Exec(os.Args[0], os.Args[0:], os.Environ())
			}
		}
	}()

	mux := xmux.New()
	c := xhandler.Chain{}
	c.UseC(xlog.NewHandler(logConf))
	c.UseC(xhandler.CloseHandler)
	c.UseC(xhandler.TimeoutHandler(10 * time.Second))
	c.UseC(xlog.RemoteAddrHandler("ip"))
	//c.UseC(xlog.UserAgentHandler("user_agent"))
	c.UseC(xlog.URLHandler("url"))
	c.UseC(xlog.RefererHandler("referer"))
	c.UseC(xlog.RequestIDHandler("req_id", "Request-Id"))
	c.UseC(logRequestC)
	c.Use(ab.ExpireMiddleware)
	c.Use(func(h http.Handler) http.Handler { return prometheus.InstrumentHandler("funder", h) })

	h := func(f func(context.Context, http.ResponseWriter, *http.Request)) xhandler.HandlerC {
		return xhandler.HandlerFuncC(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				session, err := sessionStore.Get(r, sessionCookieName)
				if err == nil {
					if len(session.Values) > 0 {
						xlog.FromContext(ctx).Debugf("found session %#v", session.Values)
					}
					ctx = context.WithValue(ctx, "session", session)
				}
				f(ctx, w, r)
			})
	}
	ha := func(f func(context.Context, http.ResponseWriter, *http.Request)) xhandler.HandlerC {
		return authProtectC(h(f))
	}

	mux.Handle("GET", "/assets/*filepath", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	abRouter := h(mkHandlerC(ab.NewRouter()).ServeHTTPC)
	mux.GET("/auth/*path", abRouter)
	mux.POST("/auth/*path", abRouter)

	admin := mux.NewGroup("/admin")
	admin.GET("/users", ha(usersGET))
	sub := admin.NewGroup("/user")
	sub.GET("/:userid", ha(userGET))
	sub.POST("/:userid", ha(userPOST))
	sub.DELETE("/:userid", ha(userDELETE))

	admin.GET("/raisers", ha(raisersGET))
	sub = admin.NewGroup("/raiser")
	sub.GET("/:raiserid", ha(raiserGET))
	sub.POST("/:raiserid", ha(raiserPOST))
	sub.DELETE("/:raiserid", ha(raiserDELETE))

	admin.GET("/funders", ha(fundersGET))
	sub = admin.NewGroup("/funder")
	sub.GET("/:funderid", ha(funderGET))
	sub.POST("/:funderid", ha(funderPOST))
	sub.DELETE("/:funderid", ha(funderDELETE))

	mux.Handle("GET", "/metrics", prometheus.Handler())
	mux.GET("/", h(rootGET))

	logger.Info("Start listening on :8080")
	logger.Fatal(http.ListenAndServe(":8080", c.Handler(mux)))
}

func rootGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var msg string
	var u User
	var ok bool
	uI, err := ab.CurrentUser(w, r)
	if err != nil {
		msg = fmt.Sprintf("CurrentUser: %v", err)
	} else if u, ok = uI.(User); ok {
		msg = fmt.Sprintf("%#v", uI)
	}
	if err := templates.Render(w, "index",
		map[string]interface{}{
			"Message":           msg,
			"loggedin":          uI != nil,
			"current_user_name": u.Name,
		}); err != nil {
		xlog.FromContext(ctx).Errorf("Render index: %v", err)
	}
}

//go:generate go run gen_keys.go
func setupAuthboss(logWriter io.Writer) {
	cookieStore = securecookie.New(cookieStoreKey, nil)
	sessionStore = sessions.NewCookieStore(sessionStoreKey)

	ab.LogWriter = os.Stderr
	ab.Storer = database
	ab.CookieStoreMaker = NewCookieStorer
	ab.SessionStoreMaker = NewSessionStorer
	ab.OAuth2Storer = database
	ab.MountPath = "/auth"
	ab.ViewsPath = "ab_views"
	ab.LayoutDataMaker = layoutData
	ab.OAuth2Providers = map[string]authboss.OAuth2Provider{
		"google": authboss.OAuth2Provider{
			OAuth2Config: &oauth2.Config{
				ClientID:     `358330097293-dniffjrngm0v2g7jb3u73dacscnq8vdj.apps.googleusercontent.com`,
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
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
	ab.Mailer = authboss.LogMailer(os.Stdout) // authboss.SMTPMailer("127.0.0.1:25")
	//ab.ErrorHandler
	ab.LogWriter = logWriter
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
	b, err := ioutil.ReadFile(filepath.Join("views", "layout.html.tpl"))
	if err != nil {
		panic(err)
	}
	ab.Layout = template.Must(template.New("layout").Funcs(funcs).Parse(string(b)))

	if err := ab.Init(); err != nil {
		xlog.Fatal(err)
	}
}

func layoutData(w http.ResponseWriter, r *http.Request) authboss.HTMLData {
	currentUserName := ""
	userInter, err := ab.CurrentUser(w, r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*User).Name
	}

	return authboss.HTMLData{
		"loggedin":               userInter != nil,
		"username":               "",
		authboss.FlashSuccessKey: ab.FlashSuccess(w, r),
		authboss.FlashErrorKey:   ab.FlashError(w, r),
		"current_user_name":      currentUserName,
	}
}

type authProtector struct {
	f xhandler.HandlerC
}

func authProtectC(f xhandler.HandlerC) xhandler.HandlerC {
	return authProtector{f}
}

func (ap authProtector) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if u, err := ab.CurrentUser(w, r); err != nil {
		log.Println("Error fetching current user:", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if u == nil {
		xlog.FromContext(ctx).Info("Redirecting unauthorized user from ", r.URL.Path)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		ap.f.ServeHTTPC(context.WithValue(ctx, "user", u), w, r)
	}
}

func writeJSON(ctx context.Context, w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		xlog.FromContext(ctx).Error(err)
	}
}

func logRequestC(h xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		logger := xlog.FromContext(ctx)
		logger.Info(r.Method)
		start := time.Now()
		lw := &logResponseWriter{ResponseWriter: w}
		h.ServeHTTPC(ctx, lw, r)
		dur := time.Since(start)
		logger.SetField("status_code", lw.StatusCode)
		logger.Debug(dur)
	})
}
func mkHandlerC(h http.Handler) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		xlog.FromContext(ctx).Debug("mkHandlerC(%v)", h)
		h.ServeHTTP(w, r)
	})
}

func StatusCodeHandler(h xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		lw := &logResponseWriter{ResponseWriter: w}
		h.ServeHTTPC(ctx, lw, r)
		xlog.FromContext(ctx).SetField("status_code", lw.StatusCode)
	})
}

type logResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lw *logResponseWriter) WriteHeader(status int) {
	lw.StatusCode = status
	lw.ResponseWriter.WriteHeader(status)
}
