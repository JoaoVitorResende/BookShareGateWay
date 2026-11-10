package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(auth *AuthHandler, books *BookHandler, mw *AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)

	// Rotas públicas
	r.Post("/auth/register", auth.Register)
	r.Post("/auth/login", auth.Login)

	// Rotas protegidas: só passam com JWT válido
	r.Group(func(r chi.Router) {
		r.Use(mw.Require)
		r.Post("/books", books.Create)
	})

	return r
}
