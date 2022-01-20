package router

import (
	"encoding/json"
	"log"
	"net/http"
	"t3/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	Id string // User unique id + remote address for unique connection

	hub *Hub

	conn *websocket.Conn

	send chan *[]models.Flight
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			text, err := json.Marshal(message)

			if err != nil {
				return
			}

			w.Write(text)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (srv *Srv) ServeWs(c *gin.Context) {

	pool := pool{
		w:  c.Writer,
		r:  c.Request,
		Ch: make(chan *websocket.Conn, 1),
	}

	// 60 second timeout for pool workers to do there job
	ticker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	srv.Hub.pool <- pool
	conn := &websocket.Conn{}

	for {
		select {
		case conn = <-pool.Ch:
		case <-ticker.C:
			log.Println("Stopped because of timeout")
			c.Writer.WriteHeader(http.StatusRequestTimeout)
			return
		}
		if conn.RemoteAddr().String() != "" {
			break
		}
	}

	client := &Client{

		// need better ID
		Id:   conn.RemoteAddr().String(),
		hub:  srv.Hub,
		conn: conn,
		send: make(chan *[]models.Flight, 10),
	}

	client.hub.register <- client

	go client.writePump()
}
