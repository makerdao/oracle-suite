package logger

import (
	"github.com/libp2p/go-libp2p-core/network"

	"github.com/makerdao/gofer/internal/log"
	"github.com/makerdao/gofer/internal/transport/p2p/sets"
)

type Node interface {
	AddNotifee(notifees ...network.Notifiee)
	AddEventHandler(eventHandler ...sets.EventHandler)
	AddMessageHandler(messageHandlers ...sets.MessageHandler)
}

// Register registers extensions to P2P node which will print additional
// debug logs.
func Register(node Node, l log.Logger) {
	node.AddNotifee(&notifee{log: l})
	node.AddEventHandler(&eventHandler{log: l})
	node.AddMessageHandler(&messageHandler{log: l})
}
