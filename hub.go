// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gowss

import (
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	Broadcast chan MsgBody

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	recv chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var MainHub *Hub

func NewHub() *Hub {
	if MainHub != nil {
		return MainHub
	}

	h := Hub{
		Broadcast:  make(chan MsgBody),
		recv:       make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	MainHub = &h
	go h.run()
	return MainHub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				client.attrs = nil
				client.lastSendMsgHash = nil
			}

		case message := <-h.Broadcast:
			go func() {
				log.Printf("broadcast: %+v", message)
				for client := range h.clients {
					//广播消息之前，检查client是否拥有对应属性
					//去重相同两条连续的重复消息
					newHash := message.BodyHash()
					if lastHash, ok := client.lastSendMsgHash[message.To]; ok {
						if newHash == lastHash {
							log.Println("消息重复跳过")
							return
						}
					}

					client.lastSendMsgHash[message.To] = message.BodyHash()

					select {
					case client.send <- message.GetBody():
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}()

		case <-h.recv:

		}
	}
}
