package stomp

import (
	"errors"
	"log"
	"net/http"
	"strings"

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

type StompFunc = func() error
type StompRequestType struct {
	Destination string
	Handler     StompFunc
}

func (s *Stomp) Send(message string, destination string) error {
	return nil
}

func (s *Stomp) Connect() error {
	ws, err := upgrader.Upgrade(s.writer, s.request, s.responseHeader)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		handleSTOMPMessage(message, ws)
	}
}

func handleSTOMPMessage(message []byte, ws *websocket.Conn) error {
	msg := string(message)
	lines := strings.Split(msg, "\n")
	if len(lines) < 2 {
		return errors.New("invalid_message")
	}
	command := lines[0]
	switch command {
	case "CONNECT", "STOMP":
		
		// Handle client connection request.
		// Typically involves authentication and acknowledging the connection.
		// Example: conn.WriteMessage(websocket.TextMessage, []byte("CONNECTED\nversion:1.2\n\n\000"))
	case "DISCONNECT":
		return ws.Close()
		// Handle client disconnection.
		// Clean up resources, close connection, etc.
	case "SUBSCRIBE":
		// Handle subscription to a destination.
		// Track subscription ID and destination for message dispatching.
	case "UNSUBSCRIBE":
		// Handle unsubscription from a destination.
		// Remove tracking of subscription ID.
	case "SEND":
		// Handle sending a message to a destination.
		// Implement logic based on the destination in the headers.
	case "BEGIN":
		// Handle the start of a transaction.
		// Track transaction ID and related actions.
	case "COMMIT":
		// Handle committing a transaction.
		// Process all actions collected under the transaction ID.
	case "ABORT":
		// Handle aborting a transaction.
		// Discard all actions collected under the transaction ID.
	case "ACK":
		// Handle acknowledgment of a message processing.
		// Typically used in conjunction with client-individual acknowledgment mode.
	case "NACK":
		// Handle negative acknowledgment of a message processing.
		// Indicate that a message was not processed successfully.
	default:
		log.Println("Unknown STOMP command:", command)
	}

	return nil

}
