package hub

import (
	"log"
)

type DefaultChannel struct {
	hub *Hub
	// Registered Clients.
	clients map[*Client]bool

	// Inbound messages from the Clients.
	broadcast chan *Message

	// Register requests from the Clients.
	register chan *Client

	// Unregister requests from Clients.
	unregister chan *Client
	params     *ChannelParam
}

type MessageDispatchFunc func(string, []byte, ...string)

func NewDefaultChannel(options ...Option) Channel {
	params := &ChannelParam{}
	for _, option := range options {
		option(params)
	}
	return &DefaultChannel{
		hub:        nil,
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		params:     params,
	}
}

func (dc *DefaultChannel) Run(hub *Hub) {
	dc.hub = hub
	for {
		select {
		case client := <-dc.register:
			dc.clients[client] = true
		case client := <-dc.unregister:
			if _, ok := dc.clients[client]; ok {
				delete(dc.clients, client)
				client.Close()
			}
			if len(dc.clients) == 0 {
				dc.hub.UnRegister(dc)
				return
			}
		case message := <-dc.broadcast:
			if dc.params.BroadcastExceptSender {
				dc.BroadcastHandler(message.ClientID, message.Content, dc.Code())
			} else {
				dc.BroadcastHandler(message.ClientID, message.Content)
			}
		}
	}
}

func (dc *DefaultChannel) BroadcastHandler(client string, message []byte, excepts ...string) {
	log.Println(string(message))
	for client := range dc.clients {
		for _, id := range excepts {
			if client.Code == id {
				goto DoNothing
			}
		}
		if err := client.Send(message); err != nil {
			delete(dc.clients, client)
		}
	DoNothing:
	}
}

func (dc *DefaultChannel) Broadcast() chan<- *Message {
	return dc.broadcast
}
func (dc *DefaultChannel) Register() chan<- *Client {
	return dc.register
}

func (dc *DefaultChannel) UnRegister() chan<- *Client {
	return dc.unregister
}

func (dc *DefaultChannel) Code() string {
	return dc.params.Code
}

func (dc *DefaultChannel) Type() string {
	return "default"
}

func (dc *DefaultChannel) remove() {

}

func (dc *DefaultChannel) TotalClients() float64 {
	return float64(len(dc.clients))
}