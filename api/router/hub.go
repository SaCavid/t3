package router

import (
	"log"
	"net/http"
	"sync"
	"t3/models"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	workers     = 256
	broadcaster = 128
)

type pool struct {
	w  http.ResponseWriter
	r  *http.Request
	Ch chan *websocket.Conn
}

type Hub struct {
	Started bool

	pool chan pool

	// write locker
	rw sync.RWMutex

	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	Broadcasts chan *[]models.Flight

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcasts: make(chan *[]models.Flight, broadcaster),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		pool:       make(chan pool, workers),
	}
}

func (h *Hub) Run() {

	for i := broadcaster; i >= 0; i-- {
		go h.Broadcast()
	}
	h.Started = true
	for {
		select {
		case client := <-h.register:
			h.rw.Lock()
			h.clients[client.Id] = client
			h.rw.Unlock()
		case client := <-h.unregister:

			h.rw.RLock()
			_, ok := h.clients[client.Id]
			h.rw.RUnlock()
			if ok {
				h.rw.Lock()
				delete(h.clients, client.Id)
				h.rw.Unlock()
				close(client.send)
			}
		}
	}
}

func (h *Hub) Broadcast() {

	for {

		select {
		case message := <-h.Broadcasts:
			log.Println("New invoke")
			h.rw.RLock()
			clients := h.clients
			h.rw.RUnlock()
			log.Println(len(clients))
			for _, v := range clients {
				v.send <- message
			}
		}
	}
}

func (h *Hub) Pool() {
	wg := &sync.WaitGroup{}
	wg.Add(workers)
	// create 256 workers for websocket registration
	for i := 0; i <= workers; i++ {
		go h.worker()
	}

	wg.Wait()
}

func (h *Hub) worker() {

	for {
		select {
		case pool := <-h.pool:
			conn, err := upgrader.Upgrade(pool.w, pool.r, nil)
			if err != nil {
				log.Println(err)
				return
			}
			pool.Ch <- conn

		}
	}
}
