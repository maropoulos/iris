// Copyright (c) 2016, Gerasimos Maropoulos
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//	  this list of conditions and the following disclaimer
//    in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse
//    or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER AND CONTRIBUTOR, GERASIMOS MAROPOULOS BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package iris v3.0.0-alpha
//
// Note: When 'Station', we mean the Iris type.
package iris

import (
	"os"

	"sync"
	"time"

	"github.com/kataras/iris/logger"
	"github.com/kataras/iris/rest"
	"github.com/kataras/iris/server"
	"github.com/kataras/iris/sessions"
	_ "github.com/kataras/iris/sessions/providers/memory"
	_ "github.com/kataras/iris/sessions/providers/redis"
	"github.com/kataras/iris/template"
	"github.com/kataras/iris/template/engine"
	"github.com/kataras/iris/template/engine/pongo"
	"github.com/kataras/iris/template/engine/standar"
)

const (
	Version = "v3.0.0-alpha"
)

//for conversional
var (
	StandarEngine = engine.Standar
	PongoEngine   = engine.Pongo
)

type (

	// SessionConfig the configuration for sessions
	// We don't import the providers and make it easier with Provider = iris.Redis OR iris.Memory [iotas] and make all the rest automatically because
	// we want to give the developers the functionality to change the options of each now/and future/or custom session provider they select
	// example: import "github.com/kataras/iris/sessions/providers/redis" ... redis.Config.Addr = "127.0.0.1:2222";iris.Config().Session.Provider = redis.Provider
	SessionConfig struct {
		// Provider string, usage iris.Config().Provider = "memory" or "redis". If you wan to customize redis then import the package, and change it's config
		Provider string
		// Secret string, the session's client cookie name, for example: "irissessionid"
		Secret string
		// Life time.Duration, cookie life duration and gc duration, for example: time.Duration(60)*time.Minute
		Life time.Duration
	}

	TemplateConfig struct {
		Engine         engine.EngineType
		*engine.Config // contains common configs for both standar & pongo

		Standar *standar.StandarConfig // contains specific configs for standar html/template
		Pongo   *pongo.PongoConfig     // contains specific configs for pongo2
	}

	// IrisConfig options for iris before server listen
	// MaxRequestBodySize is the only options that can be changed after server listen - using SetMaxRequestBodySize(int)
	// Render config can be changed after declaration but before server's listen - using Config().Render
	// Session config can be changed after declaration but before server's listen - using Config().Session
	IrisConfig struct {
		// MaxRequestBodySize Maximum request body size.
		//
		// The server rejects requests with bodies exceeding this limit.
		//
		// By default request body size is unlimited.
		MaxRequestBodySize int
		// PathCorrection corrects and redirects the requested path to the registed path
		// for example, if /home/ path is requested but no handler for this Route found,
		// then the Router checks if /home handler exists, if yes, redirects the client to the correct path /home
		// and VICE - VERSA if /home/ is registed but /home is requested then it redirects to /home/
		//
		// Default is true
		PathCorrection bool

		// Log turn it to false if you want to disable logger,
		// Iris prints/logs ONLY errors, so be careful when you disable it
		Log bool

		// Profile set to true to enable web pprof (debug profiling)
		// Default is false, enabling makes available these 7 routes:
		// /debug/pprof/cmdline
		// /debug/pprof/profile
		// /debug/pprof/symbol
		// /debug/pprof/goroutine
		// /debug/pprof/heap
		// /debug/pprof/threadcreate
		// /debug/pprof/pprof/block
		Profile bool

		// ProfilePath change it if you want other url path than the default
		// Default is /debug/pprof , which means yourhost.com/debug/pprof
		ProfilePath string

		// Template the configs for template
		Templates *TemplateConfig // inside template_config.go
		// Rest configs for rendering.
		//
		// these options inside this config don't have any relation with the TemplateEngine
		// from github.com/kataras/iris/rest
		Rest *rest.Config

		// Session the config for sessions
		// contains 3(three) properties
		// Provider: (look /sessions/providers)
		// Secret: cookie's name (string)
		// Life: cookie life (time.Duration)
		Session *SessionConfig
	}

	// Iris is the container of all, server, router, cache and the sync.Pool
	Iris struct {
		*router
		server         *server.Server
		plugins        *PluginContainer
		rest           *rest.Render
		templates      *template.Template
		sessionManager *sessions.Manager

		config *IrisConfig
		logger *logger.Logger
	}
)

// New creates and returns a new iris Iris. If config is empty then default config is used
//
// Receives an optional iris.IrisConfig as parameter
// If empty then iris.DefaultConfig() are used
func New(configs ...*IrisConfig) *Iris {
	config := DefaultConfig()
	if configs != nil && len(configs) > 0 {
		config = configs[0]
	}

	if config.ProfilePath == "" {
		config.ProfilePath = DefaultProfilePath
	}

	// create the Iris
	s := &Iris{config: config, plugins: &PluginContainer{}}

	// create & set the router
	s.router = newRouter(s)

	// set the Logger
	s.logger = logger.New()

	return s
}

// Server returns the server
func (s *Iris) Server() *server.Server {
	return s.server
}

// Plugins returns the plugin container
func (s *Iris) Plugins() *PluginContainer {
	return s.plugins
}

// Config returns the configs
func (s *Iris) Config() *IrisConfig {
	return s.config
}

// Logger returns the logger
func (s *Iris) Logger() *logger.Logger {
	return s.logger
}

// Render returns the rest render
func (s *Iris) Rest() *rest.Render {
	return s.rest
}

// Templates returns the template render
func (s *Iris) Templates() *template.Template {
	return s.templates
}

// SetMaxRequestBodySize Maximum request body size.
//
// The server rejects requests with bodies exceeding this limit.
//
// By default request body size is unlimited.
func (s *Iris) SetMaxRequestBodySize(size int) {
	s.config.MaxRequestBodySize = size
}

// newContextPool returns a new context pool, internal method used in tree and router
func (s *Iris) newContextPool() sync.Pool {
	return sync.Pool{New: func() interface{} {
		return &Context{station: s}
	}}
}

// DoPreListen call router's optimize, sets the server's handler and notice the plugins
// receives the server.Config
// returns the station's Server (*server.Server)
// it's a non-blocking func
func (s *Iris) DoPreListen(opt server.Config) *server.Server {
	//runs only once even if called more than one time.

	// set the logger's state
	s.logger.SetEnable(s.config.Log)

	// set the rest render (for Data, Text, JSON, JSONP, XML)
	s.rest = rest.New(s.config.Rest)

	// determinate which template engine is used and set the template wrapper (for html or whatever extension was given)
	var e engine.Engine

	ct := s.config.Templates

	switch s.config.Templates.Engine {
	case engine.Pongo:
		e = pongo.New(pongo.WrapConfig(ct.Config, ct.Pongo))
	default:
		e = standar.New(standar.WrapConfig(ct.Config, ct.Standar)) // default to standar
	}
	// I could also do a  check if Pongo's config != empty then use pongo2 but this will brings unexpecting results because the user must explicit give which engine wants via the Engine field
	s.templates = template.New(e)

	// router prepare
	if !s.router.optimized {
		s.router.optimize()

		s.server = server.New(opt)
		s.server.SetHandler(s.router.ServeRequest)

		if s.config.MaxRequestBodySize > 0 {
			s.server.MaxRequestBodySize = s.config.MaxRequestBodySize
		}
	}

	s.plugins.DoPreListen(s)

	return s.server
}

// DoPostListen sets the render and notice the plugins
// it's a non-blocking func
func (s *Iris) DoPostListen() {

	if s.config.Session != nil && s.config.Session.Provider != "" {
		if s.config.Session.Secret == "" {
			s.config.Session.Secret = DefaultCookieName
		}
		if s.config.Session.Life == 0 {
			s.config.Session.Life = DefaultCookieDuration
		}
		s.sessionManager = sessions.New(s.config.Session.Provider, s.config.Session.Secret, s.config.Session.Life)
	}

	s.plugins.DoPostListen(s)
}

// openServer is internal method, open the server with specific options passed by the Listen and ListenTLS
// it's a blocking func
func (s *Iris) openServer(opt server.Config) (err error) {
	s.DoPreListen(opt)

	if err = s.server.OpenServer(); err == nil {
		s.DoPostListen()
		ch := make(chan os.Signal)
		<-ch
		s.Close()
	}
	return
}

// Listen starts the standalone http server
// which listens to the addr parameter which as the form of
// host:port or just port
//
// It panics on error if you need a func to return an error use the ListenWithErr
// ex: iris.Listen(":8080")
func (s *Iris) Listen(addr string) {
	opt := server.Config{ListeningAddr: addr}
	if err := s.openServer(opt); err != nil {
		panic(err)
	}
}

// ListenWithErr starts the standalone http server
// which listens to the addr parameter which as the form of
// host:port or just port
//
// It returns an error you are responsible how to handle this
// if you need a func to panic on error use the Listen
// ex: log.Fatal(iris.ListenWithErr(":8080"))
func (s *Iris) ListenWithErr(addr string) error {
	opt := server.Config{ListeningAddr: addr}
	return s.openServer(opt)
}

// ListenTLS Starts a https server with certificates,
// if you use this method the requests of the form of 'http://' will fail
// only https:// connections are allowed
// which listens to the addr parameter which as the form of
// host:port or just port
//
// It panics on error if you need a func to return an error use the ListenTLSWithErr
// ex: iris.ListenTLS(":8080","yourfile.cert","yourfile.key")
func (s *Iris) ListenTLS(addr string, certFile, keyFile string) {
	opt := server.Config{ListeningAddr: addr, CertFile: certFile, KeyFile: keyFile}
	if err := s.openServer(opt); err != nil {
		panic(err)
	}
}

// ListenTLSWithErr Starts a https server with certificates,
// if you use this method the requests of the form of 'http://' will fail
// only https:// connections are allowed
// which listens to the addr parameter which as the form of
// host:port or just port
//
// It returns an error you are responsible how to handle this
// if you need a func to panic on error use the ListenTLS
// ex: log.Fatal(iris.ListenTLSWithErr(":8080","yourfile.cert","yourfile.key"))
func (s *Iris) ListenTLSWithErr(addr string, certFile, keyFile string) error {
	opt := server.Config{ListeningAddr: addr, CertFile: certFile, KeyFile: keyFile}
	return s.openServer(opt)
}

// Close is used to close the tcp listener from the server
func (s *Iris) Close() error {
	s.plugins.DoPreClose(s)
	return s.server.CloseServer()
}
