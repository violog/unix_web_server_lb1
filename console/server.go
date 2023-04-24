// Copyright (C) 2021 Creditor Corp. Group.
// See LICENSE for copying information.

package console

import (
	"context"
	"errors"
	"html/template"
	"net"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"todo/console/controllers"
	"todo/items"
	"todo/pkg/auth"
	"todo/users"
	"todo/users/userauth"
)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("admin web server error")
)

// Config contains configuration for admin web server.
type Config struct {
	Address   string `json:"address"`
	StaticDir string `json:"staticDir"`

	Auth struct {
		CookieName string `json:"cookieName"`
		Path       string `json:"path"`
	} `json:"auth"`
}

// AbsolutePath defines path to html templates.
const AbsolutePath = "/Users/macos/go/src/awesomeProject/todo/web"

// Server represents admin web server.
//
// architecture: Endpoint
type Server struct {
	log *zap.Logger

	listener net.Listener
	server   http.Server

	authService *userauth.Service
	cookieAuth  *auth.CookieAuth

	templates struct {
		items controllers.ItemsTemplates
		auth  controllers.AuthTemplates
	}
}

// NewServer is a constructor for admin web server.
func NewServer(listener net.Listener, authService *userauth.Service, logger *zap.Logger, items *items.Service, users *users.Service) (*Server, error) {
	server := &Server{
		cookieAuth: auth.NewCookieAuth(auth.CookieSettings{
			Name: "todo",
			Path: "/",
		}),
		authService: authService,
		listener:    listener,
		log:         logger,
	}

	err := server.initializeTemplates()
	if err != nil {
		server.log.Error(err.Error())
		return server, err
	}

	router := mux.NewRouter()
	authController := controllers.NewAuth(server.log, server.authService, server.cookieAuth, server.templates.auth, users)
	router.HandleFunc("/login", authController.Login).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register", authController.Register).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/logout", authController.Logout).Methods(http.MethodGet)

	itemsRouter := router.PathPrefix("/{userId}/items").Subrouter()
	itemsRouter.Use(server.withAuth)
	itemsController := controllers.NewItems(server.log, items, server.templates.items)
	itemsRouter.HandleFunc("", itemsController.List).Methods(http.MethodGet)
	itemsRouter.HandleFunc("/create", itemsController.Create).Methods(http.MethodGet, http.MethodPost)
	itemsRouter.HandleFunc("/update/{id}", itemsController.Update).Methods(http.MethodGet, http.MethodPost)
	itemsRouter.HandleFunc("/update-status/{id}", itemsController.UpdateStatus).Methods(http.MethodGet, http.MethodPost)
	itemsRouter.HandleFunc("/delete/{id}", itemsController.Delete).Methods(http.MethodGet)

	server.server = http.Server{
		Handler: router,
	}

	return server, nil
}

// Run starts the server that host webapp and api endpoint.
func (server *Server) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		return server.server.Shutdown(context.Background())
	})
	group.Go(func() error {
		defer cancel()
		err := server.server.Serve(server.listener)
		isCancelled := errs.IsFunc(err, func(err error) bool { return errors.Is(err, context.Canceled) })
		if isCancelled || errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	})

	return group.Wait()
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return server.server.Close()
}

// withAuth performs initial authorization before every request.
func (server *Server) withAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context

		ctxWithAuth := func(ctx context.Context) context.Context {
			token, err := server.cookieAuth.GetToken(r)
			if err != nil {
				controllers.Redirect(w, r, "/login", http.MethodGet)
			}

			claims, err := server.authService.Authorize(ctx, token)
			if err != nil {
				controllers.Redirect(w, r, "/login", http.MethodGet)
			}

			return auth.SetClaims(ctx, claims)
		}

		ctx = ctxWithAuth(r.Context())

		handler.ServeHTTP(w, r.Clone(ctx))
	})
}

// initializeTemplates initializes and caches templates for managers controller.
func (server *Server) initializeTemplates() (err error) {
	server.templates.auth.Login, err = template.ParseFiles(filepath.Join("web", "auth", "login.html"))
	if err != nil {
		return err
	}
	server.templates.auth.Register, err = template.ParseFiles(filepath.Join("web", "auth", "register.html"))
	if err != nil {
		return err
	}

	server.templates.items.List, err = template.ParseFiles(filepath.Join("web", "items", "list.html"))
	if err != nil {
		return err
	}
	server.templates.items.Create, err = template.ParseFiles(filepath.Join("web", "items", "create.html"))
	if err != nil {
		return err
	}
	server.templates.items.Update, err = template.ParseFiles(filepath.Join("web", "items", "update.html"))
	if err != nil {
		return err
	}

	return nil
}
