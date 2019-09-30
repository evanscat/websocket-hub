package hub

import (
	"encoding/json"
	"log"
)

type MultiMessage struct {
	Dest    string
	Src     string
	Type    string
	Content json.RawMessage
}

type Grouper interface {
	FindMembers(group string) (map[string]struct{}, bool)
	HandleRegister(client *Client)
	HandleUnRegister(client *Client)
	Add(clientID, groupID string)
	Remove(clientID, groupID string)
}

type DefaultGrouper struct {
	router map[string]map[string]struct{}
}

type MultiChannel struct {
	// Registered Clients.
	Clients map[string]*Client
	// Inbound messages from the Clients.
	broadcast chan *Message

	// Register requests from the Clients.
	register chan *Client

	// Unregister requests from Clients.
	unregister chan *Client
	params     *ChannelParam
	grouper    Grouper
}

func NewMultiChannel(options ...Option) Channel {
	params := &ChannelParam{}
	for _, option := range options {
		option(params)
	}
	return &MultiChannel{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		params:     params,
		grouper:    &DefaultGrouper{router: make(map[string]map[string]struct{})},
	}
}

func (cc *MultiChannel) Run(hub *Hub) {
	for {
		select {
		case client := <-cc.register:
			if cli, ok := cc.Clients[client.Code]; ok {
				cli.Close()
			}
			cc.Clients[client.Code] = client
			cc.grouper.HandleRegister(client)
		case client := <-cc.unregister:
			if cli, ok := cc.Clients[client.Code]; ok && cli == client {
				delete(cc.Clients, client.Code)
				client.Close()
				cc.grouper.HandleUnRegister(client)
			}
			if len(cc.Clients) == 0 {
				hub.UnRegister(cc)
				return
			}
		case message := <-cc.broadcast:
			if cc.params.BroadcastExceptSender {
				cc.BroadcastHandler(message.ClientID, message.Content, cc.Code())
			} else {
				cc.BroadcastHandler(message.ClientID, message.Content)
			}
		}
	}
}

func (cc *MultiChannel) BroadcastHandler(client string, message []byte, excepts ...string) {
	msg := &MultiMessage{}
	err := json.Unmarshal(message, msg)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(message))
	//没有该分组
	subs, ok := cc.grouper.FindMembers(msg.Dest)
	if !ok {
		return
	}
	//不在分组内
	if cc.params.CheckSenderAuthority {
		if _, ok := subs[msg.Src]; !ok {
			return
		}
	}
	for key := range subs {
		//except不发送
		for _, id := range excepts {
			if client == id {
				goto DoNothing
			}
		}
		//取出active链接发送
		if cli, ok := cc.Clients[key]; ok {
			if err = cli.Send(message); err != nil {
				delete(cc.Clients, cli.Code)
			}
		}
	DoNothing:
	}
}

func (cc *MultiChannel) Broadcast() chan<- *Message {
	return cc.broadcast
}
func (cc *MultiChannel) Register() chan<- *Client {
	return cc.register
}

func (cc *MultiChannel) UnRegister() chan<- *Client {
	return cc.unregister
}

func (cc *MultiChannel) Code() string {
	return cc.params.Code
}

func (cc *MultiChannel) Type() string {
	return "multi"
}

func (cc *MultiChannel) TotalClients() float64 {
	return float64(len(cc.Clients))
}

func (dg *DefaultGrouper) FindMembers(group string) (map[string]struct{}, bool) {
	ret, ok := dg.router[group]
	return ret, ok
}

func (dg *DefaultGrouper) Remove(clientID, groupID string) {
	delete(dg.router[groupID], clientID)
}

func (dg *DefaultGrouper) Add(clientID, groupID string) {
	dg.router[groupID][clientID] = struct{}{}
}

func (dg *DefaultGrouper) HandleRegister(client *Client) {
	return
}

func (dg *DefaultGrouper) HandleUnRegister(client *Client) {
	return
}

func (cc *MultiChannel) SetGroupHandler(grouper Grouper) {
	cc.grouper = grouper
}
