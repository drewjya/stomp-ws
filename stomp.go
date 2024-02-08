package stomp

import (
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/drewjya/stomp-ws/lib/subscription"
	"github.com/gorilla/websocket"
)

type Map map[string]interface{}

type Config struct {
	Destinations []string
}

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
	config         Config
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

type Handler = func(request []byte) (string, []byte, error)

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
		handleSTOMPMessage(stompHandler{Message: message, Ws: ws, Manager: manager}, handler, s.config.Destinations)
	}
}

func handleSTOMPMessage(sh stompHandler, handle Handler, destinations []string) error {

	msg := string(sh.Message)
	lines := strings.Split(msg, "\n")
	if len(lines) < 2 {
		return errors.New("invalid_message")
	}
	command := lines[0]
	switch command {
	case "CONNECT", "STOMP":
		sh.Ws.WriteMessage(200, []byte("CONNECTED\nversion:1.2\n\n\000"))

		// Handle client connection request.
		// Typically involves authentication and acknowledging the connection.
		// Example: conn.WriteMessage(websocket.TextMessage, []byte("CONNECTED\nversion:1.2\n\n\000"))
	case "DISCONNECT":
		sh.Manager.RemoveAllSubscriptions(sh.Ws)
		return sh.Ws.Close()
		// Handle client disconnection.
		// Clean up resources, close connection, etc.
	case "SUBSCRIBE":
		destination := ""
		for _, v := range lines {
			if strings.HasPrefix(v, "destination:") {

				destination = strings.TrimPrefix(v, "destination:")
				if !slices.Contains(destinations, destination) {
					return errors.New("invalid_destination")
				}
				break
			}
		}
		if destination == "" {
			return errors.New("invalid_destination")
		}
		sh.Manager.AddSubscription(destination, sh.Ws)

	case "UNSUBSCRIBE":
		for _, v := range lines {
			if strings.HasPrefix(v, "id:") {
				subscriptionId := strings.TrimPrefix(v, "id:")
				sh.Manager.RemoveSubscription(subscriptionId, sh.Ws)
				break
			}
		}

	case "SEND":
		destination, response, err := handle(sh.Message)
		if err != nil {
			return err
		}
		if slices.Contains(destinations, destination) {
			resp := []byte("MESSAGE\n")
			resp = append(resp, response...)
			resp = append(resp, []byte("\n\n\000")...)
			sh.Manager.SendMessageToSubscribers(destination, resp)
		} else {
			return errors.New("invalid_destination")

		}

	case "BEGIN":
		sh.Ws.WriteMessage(200, []byte("ACK\n\n\000"))
		// Handle the start of a transaction.
		// Track transaction ID and related actions.
	case "COMMIT":
		sh.Ws.WriteMessage(200, []byte("COMMIT\n\n\000"))
		// Handle committing a transaction.
		// Process all actions collected under the transaction ID.
	case "ABORT":
		sh.Ws.WriteMessage(200, []byte("ABORT\n\n\000"))
		// Handle aborting a transaction.
		// Discard all actions collected under the transaction ID.
	case "ACK":
		sh.Ws.WriteMessage(200, []byte("ACK\n\n\000"))
		// Handle acknowledgment of a message processing.
		// Typically used in conjunction with client-individual acknowledgment mode.
	case "NACK":
		sh.Ws.WriteMessage(200, []byte("NACK\n\n\000"))
		// Handle negative acknowledgment of a message processing.
		// Indicate that a message was not processed successfully.
	default:
		log.Println("Unknown STOMP command:", command)
	}

	return nil

}
