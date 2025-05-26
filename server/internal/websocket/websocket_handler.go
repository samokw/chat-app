package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/internal/token"
	"server/internal/user"
	"server/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *Hub
	jwtMaker *token.JWTMaker
	user.Repository
}

func NewHandler(hub *Hub, jwtMaker *token.JWTMaker, service *user.Service) *Handler {
	return &Handler{
		hub:      hub,
		jwtMaker: jwtMaker,
		service:  service,
	}
}

type CreateRoomReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid payload", err)
		return
	}
	h.hub.Rooms[req.ID] = &Room{
		ID:      req.ID,
		Name:    req.Name,
		Clients: make(map[string]*Client),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
	log.Println("created a room")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// CheckOrigin: func(r *http.Request) bool {
	// 	origin := r.Header.Get("Origin")
	// 	return origin == "http:localhost:3000"
	// },
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {

	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "room ID required", nil)
		return
	}
	userID, username, err := h.getUserFromToken(r)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "authenication required", err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "error creating room", err)
		return
	}
	client := &Client{
		Conn: conn,
		Message: make(chan *Message, 25),
		ID: userID,
		RoomID: roomID,
		Username: username,
	}
	m := &Message{
		Content: fmt.Sprintf("%s has joined the room", username),
		RoomID: roomID,
		Username: username,
	}

	// Register a client 
	h.hub.Register <- client

	// Brodcast the message

	// writeMessage()

	// readMessage()
}

func (h *Handler) getUserFromToken(r *http.Request) (userID, username string, err error) {
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		return "", "", err
	}
	tokenString := tokenCookie.Value
	claims, err := h.jwtMaker.VerifyToken(tokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}
	ctx := r.Context()
	user, err := h.Repository.GetUserByEmail(ctx, claims.Email)
	if err != nil {
		return "", "", fmt.Errorf("unable to find username: %w", err)
	}

	return strconv.Itoa(int(user.ID)), user.Username, nil
}
