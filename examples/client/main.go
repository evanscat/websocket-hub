package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/evanscat/websocket-hub"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type MyMessage struct {
	Text string
}

func Send(conn *websocket.Conn, contents ...string) {
	if len(contents) == 1 {
		conn.WriteMessage(websocket.TextMessage, []byte(contents[0]))
	} else if len(contents) == 2 {
		bt, _ := json.Marshal(&struct {
			Body string
		}{Body: contents[1]})
		msg := &hub.MultiMessage{Dest: contents[0], Src: "", Type: "chat", Content: bt}
		conn.WriteJSON(msg)
	}
}

func main() {
	dial := websocket.Dialer{}
	var channel string
	var tp string
	flag.StringVar(&channel, "channel", "test", "")
	flag.StringVar(&tp, "type", "default", "")
	flag.Parse()
	conn, resp, err := dial.Dial(fmt.Sprintf("ws://127.0.0.1:8196/ws?channel=%s&type=%s", channel, tp), http.Header{})
	if err != nil {
		log.Println(err, resp)
		return
	}
	log.Println(resp)
	go func() {
		for {
			var input string
			var des string
			ln, err := fmt.Scan(&des, &input)
			if err != nil {
				log.Println(err, ln)
			}
			log.Println(input, des)
			Send(conn, des, input)
		}
	}()

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}
		log.Println(t, string(msg))
	}
}
