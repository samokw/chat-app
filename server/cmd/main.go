package main

import (
	"log"
	"net/http"
	"server/db"
	"server/internal/routes"
	"server/internal/user"
	"server/internal/websocket"

	"github.com/ianschenck/envflag"
)

const minSecretKeySize = 32

func main() {
	var secretKey = envflag.String("SECRET_KEY", "0123456789012345678901234567890123456789", "secret key for jwt signing")
	if len(*secretKey) < minSecretKeySize{
		log.Fatalf("SECRET_KEY must be at least %d characters", minSecretKeySize)
	}

	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	userRepo := user.NewRepository(dbConn.GetDB())
	//jwtMaker := token.NewJwtMaker("your-very-secret-key")
	userService := user.NewService(userRepo, *secretKey)
	userHandler := user.NewHandler(userService)

	hub := websocket.NewHub()
	websocketHandler := websocket.NewHandler(hub)
	r := routes.InitRouter(userHandler, websocketHandler)

	http.ListenAndServe(":8080", r)
}
