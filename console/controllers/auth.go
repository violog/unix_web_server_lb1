package controllers

import (
	"html/template"
	"net/http"

	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"todo/pkg/auth"
	"todo/users"
	"todo/users/userauth"
)

// AuthError is a internal error for auth controller.
var AuthError = errs.Class("auth controller error")

// AuthTemplates holds all auth related templates.
type AuthTemplates struct {
	Login    *template.Template
	Register *template.Template
}

// Auth login authentication entity.
type Auth struct {
	log     *zap.Logger
	service *userauth.Service
	cookie  *auth.CookieAuth

	users *users.Service

	templates AuthTemplates
}

// NewAuth returns new instance of Auth.
func NewAuth(log *zap.Logger, service *userauth.Service, authCookie *auth.CookieAuth, templates AuthTemplates, users *users.Service) *Auth {
	return &Auth{
		log:       log,
		service:   service,
		cookie:    authCookie,
		users:     users,
		templates: templates,
	}
}

// Login is an endpoint to authorize admin and set auth cookie in browser.
func (auth *Auth) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	switch r.Method {
	case http.MethodGet:
		if err = auth.templates.Login.Execute(w, nil); err != nil {
			auth.log.Error("could not execute login template" + AuthError.Wrap(err).Error())
			http.Error(w, "could not execute login template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err = r.ParseForm(); err != nil {
			http.Error(w, "could not parse login form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		// TODO: change chek
		if len(email) == 0 || len(password) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		response, err := auth.service.Token(ctx, email, password)
		if err != nil {
			auth.log.Error("could not get auth token " + AuthError.Wrap(err).Error())
			switch {
			case users.ErrNoUser.Has(err):
				http.Error(w, "could not get auth token", http.StatusNotFound)
			case userauth.ErrUnauthenticated.Has(err):
				http.Error(w, "could not get auth token", http.StatusUnauthorized)
			default:
				http.Error(w, "could not get auth token", http.StatusInternalServerError)
			}

			return
		}

		auth.cookie.SetTokenCookie(w, response)

		user, err := auth.users.GetByEmail(ctx, email)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		Redirect(w, r, "/"+user.ID.String()+"/items", http.MethodGet)
	}
}

// Register is an endpoint to register new user in system.
func (auth *Auth) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error

	switch r.Method {
	case http.MethodGet:
		if err = auth.templates.Register.Execute(w, nil); err != nil {
			auth.log.Error("could not execute register template " + AuthError.Wrap(err).Error())
			http.Error(w, "could not execute register template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err = r.ParseForm(); err != nil {
			http.Error(w, "could not parse register form", http.StatusBadRequest)
			return
		}

		email := r.Form["email"]
		password := r.Form["password"]
		// TODO: change check
		if len(email) == 0 || len(password) == 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if email[0] == "" || password[0] == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := auth.users.Create(ctx, email[0], password[0])
		if err != nil {
			auth.log.Error("could not create user " + AuthError.Wrap(err).Error())
			http.Error(w, "could not create user ", http.StatusInternalServerError)
			return
		}

		Redirect(w, r, "/login", http.MethodGet)
	}
}

// Logout is an endpoint to log out and remove auth cookie from browser.
func (auth *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	auth.cookie.RemoveTokenCookie(w)
	Redirect(w, r, "/login", "GET")
}
