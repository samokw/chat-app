package routes

import (
	"server/internal/user"
	"server/internal/websocket"

	"github.com/go-chi/chi/v5"
)

func InitRouter(userHandler *user.Handler, websocketHandler *websocket.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/signup", userHandler.CreateUser)
	r.Post("/login", userHandler.Login)
	r.Post("/refresh", userHandler.RefreshToken)
	r.Post("/logout", userHandler.Logout)


	r.Post("/websocket/createRoom", websocketHandler.CreateRoom)
	return r
}
