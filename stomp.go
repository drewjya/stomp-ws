package stomp

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/drewjya/stomp-ws/lib/subscription"
	"github.com/gorilla/websocket"
)

type Map map[string]interface{}

var (
	upgrader = websocket.Upgrader{}
)

func New(w http.ResponseWriter, r *http.Request, responseHeader http.Header) *Stomp {
	return &Stomp{
		writer:         w,
		request:        r,
		responseHeader: responseHeader,
	}
}

type Stomp struct {
	writer         http.ResponseWriter
	request        *http.Request
	responseHeader http.Header
}

type WsStompFunc = func() error
type WsStompRequestType struct {
	Destination string
	Handler     WsStompFunc
}

type stompHandler struct {
	Message []byte
	Ws      *websocket.Conn
	Manager *subscription.SubscriptionManager
}

type Handler = func([]byte, *subscription.SubscriptionManager) error

func (s *Stomp) Send(message string, destination string) error {
	return nil
}

func (s *Stomp) Connect(handler Handler) error {
	ws, err := upgrader.Upgrade(s.writer, s.request, s.responseHeader)
	manager := subscription.NewSubscriptionManager()
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		handleSTOMPMessage(stompHandler{Message: message, Ws: ws, Manager: manager}, handler)
	}
}

func handleSTOMPMessage(handler stompHandler, handle Handler) error {
	msg := string(handler.Message)
	lines := strings.Split(msg, "\n")
	if len(lines) < 2 {
		return errors.New("invalid_message")
	}
	command := lines[0]
	switch command {
	case "CONNECT", "STOMP":
		handler.Ws.WriteMessage(200, []byte("CONNECTED\nversion:1.2\n\n\000"))

		// Handle client connection request.
		// Typically involves authentication and acknowledging the connection.
		// Example: conn.WriteMessage(websocket.TextMessage, []byte("CONNECTED\nversion:1.2\n\n\000"))
	case "DISCONNECT":
		handler.Manager.RemoveAllSubscriptions(handler.Ws)
		return handler.Ws.Close()
		// Handle client disconnection.
		// Clean up resources, close connection, etc.
	case "SUBSCRIBE":
		destination := ""
		for _, v := range lines {
			if strings.HasPrefix(v, "destination:") {
				destination = strings.TrimPrefix(v, "destination:")
				break
			}
		}
		if destination == "" {
			return errors.New("invalid_destination")
		}
		handler.Manager.AddSubscription(destination, handler.Ws)

	case "UNSUBSCRIBE":
		for _, v := range lines {
			if strings.HasPrefix(v, "id:") {
				subscriptionId := strings.TrimPrefix(v, "id:")
				handler.Manager.RemoveSubscription(subscriptionId, handler.Ws)
				break
			}
		}

	case "SEND":
		return handle(handler.Message, handler.Manager)

	case "BEGIN":
		handler.Ws.WriteMessage(200, []byte("ACK\n\n\000"))
		// Handle the start of a transaction.
		// Track transaction ID and related actions.
	case "COMMIT":
		handler.Ws.WriteMessage(200, []byte("COMMIT\n\n\000"))
		// Handle committing a transaction.
		// Process all actions collected under the transaction ID.
	case "ABORT":
		handler.Ws.WriteMessage(200, []byte("ABORT\n\n\000"))
		// Handle aborting a transaction.
		// Discard all actions collected under the transaction ID.
	case "ACK":
		handler.Ws.WriteMessage(200, []byte("ACK\n\n\000"))
		// Handle acknowledgment of a message processing.
		// Typically used in conjunction with client-individual acknowledgment mode.
	case "NACK":
		handler.Ws.WriteMessage(200, []byte("NACK\n\n\000"))
		// Handle negative acknowledgment of a message processing.
		// Indicate that a message was not processed successfully.
	default:
		log.Println("Unknown STOMP command:", command)
	}

	return nil

}
