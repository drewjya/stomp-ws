package subscription

import (
	"log"

	"github.com/gorilla/websocket"
)

type Subscription struct {
    conn *websocket.Conn
    topic string
}

// SubscriptionManager to hold subscriptions.
type SubscriptionManager struct {
    // Map of topic to list of subscriptions.
    subscriptions map[string][]*Subscription
}

// NewSubscriptionManager creates a new instance of SubscriptionManager.
func NewSubscriptionManager() *SubscriptionManager {
    return &SubscriptionManager{
        subscriptions: make(map[string][]*Subscription),
    }
}

// AddSubscription adds a new subscription for a topic.
func (sm *SubscriptionManager) AddSubscription(topic string, conn *websocket.Conn) {
    sm.subscriptions[topic] = append(sm.subscriptions[topic], &Subscription{conn, topic})
}

func (sm *SubscriptionManager) RemoveAllSubscriptions(conn *websocket.Conn) {
    for topic, subs := range sm.subscriptions {
        for i, sub := range subs {
            if sub.conn == conn {
                sm.subscriptions[topic] = append(subs[:i], subs[i+1:]...)
                break
            }
        }
    }
}

// RemoveSubscription removes a subscription for a topic.
func (sm *SubscriptionManager) RemoveSubscription(topic string, conn *websocket.Conn) {
    subs, ok := sm.subscriptions[topic]
    if !ok {
        return
    }
    for i, sub := range subs {
        if sub.conn == conn {
            sm.subscriptions[topic] = append(subs[:i], subs[i+1:]...)
            break
        }
    }
}

// GetSubscriptions returns all subscriptions for a topic.
func (sm *SubscriptionManager) GetSubscriptions(topic string) []*Subscription {
    return sm.subscriptions[topic]
}

// SendMessageToSubscribers sends a message to all subscribers of a topic.
func (sm *SubscriptionManager) SendMessageToSubscribers(topic string, message []byte) {
    for _, sub := range sm.GetSubscriptions(topic) {
        if err := sub.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            log.Println("Error sending message:", err)
        }
    }
}
