package main

import (
	"github.com/evanscat/websocket-hub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	h := hub.New()
	//res := prometheus.NewRegistry()
	//res := prometheus.DefaultGatherer
	hub.Factory().Register("self",NewMyChannel)
	prometheus.MustRegister(h)
	http.Handle("/ws", h)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8196", nil)

}

type Grouper struct {
	router  map[string]struct{}
}

func (dg *Grouper) FindMembers(group string) (map[string]struct{}, bool) {
	return dg.router,true
}

func (dg *Grouper) Remove(clientID, groupID string) {
	return
}

func (dg *Grouper) Add(clientID, groupID string) {
	return
}


func (dg *Grouper) HandleRegister(client *hub.Client) {
	dg.router[client.Code] = struct{}{}
	return
}

func (dg *Grouper) HandleUnRegister(client *hub.Client) {
	return
}

func NewMyChannel(opts ...hub.Option) hub.Channel {
	h := hub.NewMultiChannel(opts...)
	h.(*hub.MultiChannel).SetGroupHandler(&Grouper{make(map[string]struct{})})
	return h
}