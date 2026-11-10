package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/satellite-communications/gateway/internal/events"
)

func handleWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	broadcaster *events.Broadcaster,
) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*"},
	})
	if err != nil {
		log.Printf("WebSocket connection failed: %v", err)
		return
	}
	defer conn.CloseNow()

	log.Println("WebSocket client connected")

	// Subscribe this browser to gateway events.
	subscription := broadcaster.Subscribe()
	defer broadcaster.Unsubscribe(subscription)

	// One context for the lifetime of this connection.
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// One reader goroutine per WebSocket connection.
	go func() {
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				cancel()
				return
			}
		}
	}()

	// Forward gateway events to this browser.
	for {
		select {
		case event, ok := <-subscription:
			if !ok {
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal gateway event: %v", err)
				continue
			}

			writeCtx, writeCancel := context.WithTimeout(
				ctx,
				5*time.Second,
			)

			err = conn.Write(writeCtx, websocket.MessageText, data)
			writeCancel()

			if err != nil {
				log.Printf("WebSocket write failed: %v", err)
				return
			}

		case <-ctx.Done():
			log.Println("WebSocket client disconnected")
			return
		}
	}
}
