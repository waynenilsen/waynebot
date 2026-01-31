package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

const wsTicketExpiry = 30 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Origin checking is handled by CORS middleware on the ticket endpoint.
		// The WebSocket upgrade itself doesn't carry cookies/auth, just the ticket.
		return true
	},
}

// WsHandler handles WebSocket ticket creation and connection upgrades.
type WsHandler struct {
	DB  *db.DB
	Hub *ws.Hub
}

type wsTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt string `json:"expires_at"`
}

// CreateTicket mints a single-use WebSocket ticket for the authenticated user.
// POST /api/ws/ticket
func (h *WsHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ticket, err := auth.GenerateToken()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	expiresAt := time.Now().Add(wsTicketExpiry)
	t, err := model.CreateWsTicket(h.DB, ticket, user.ID, expiresAt)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusCreated, wsTicketResponse{
		Ticket:    t.Ticket,
		ExpiresAt: t.ExpiresAt.Format(time.RFC3339),
	})
}

// Upgrade handles the WebSocket upgrade via ticket-based auth.
// GET /ws?ticket=...
func (h *WsHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "missing ticket", http.StatusBadRequest)
		return
	}

	t, err := model.ClaimWsTicket(h.DB, ticket)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid or expired ticket", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	client := ws.NewClient(h.Hub, conn, t.UserID)
	h.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
