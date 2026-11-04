package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(auth *AuthHandler, mw *AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)

	r.Post("/auth/register", auth.Register)
	r.Post("/auth/login", auth.Login)

	r.Group(func(r chi.Router) {
		r.Use(mw.Require)
		// rotas protegidas de livros entram aqui
	})

	return r
}