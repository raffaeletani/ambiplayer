package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebsocketReturn struct {
	Lib  []WebsocketLibrary
	Play []WebsocketStatus
}

type WebsocketLibrary struct {
	Name            string
	Code            string
	DefaultFadeTime int
}
type WebsocketStatus struct {
	Id              int
	Name            string
	Code            string
	Seek            Seek
	Done            bool
	Paused          bool
	DefaultFadeTime int
	Volume          int
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
	},
}

type WebsocketServer struct {
	clients map[*websocket.Conn]bool
	//handleMessage func(message []byte) // New message handler
}

var server WebsocketServer

func CreateWebsocketServer(router *gin.Engine) {
	server = WebsocketServer{
		make(map[*websocket.Conn]bool),
	}
	router.GET("/ws", server.playstate)

}

func (server *WebsocketServer) playstate(c *gin.Context) {
	connection, _ := upgrader.Upgrade(c.Writer, c.Request, nil)

	server.clients[connection] = true // Save the connection using it as a key

	for {
		mt, _, err := connection.ReadMessage()

		if err != nil || mt == websocket.CloseMessage {
			break // Exit the loop if the client tries to close the connection or the connection is interrupted
		}

		go server.CalcProgres()

	}
	fmt.Println("Ws connection closed")
	delete(server.clients, connection) // Removing the connection

	connection.Close()
}

func (server *WebsocketServer) CalcProgres() {
	wsret := new(WebsocketReturn)
	var audiofile *AudioFile
	var status WebsocketStatus
	bla := Playing //make copy to read without interfering with the rest
	for idx := range bla {
		audiofile = bla[idx]
		if audiofile.Seeker != nil {
			audiofile.Seek.PosDuration = audiofile.Format.SampleRate.D(audiofile.Seeker.Position())
			audiofile.Seek.LenDuration = audiofile.Format.SampleRate.D(audiofile.Seeker.Len())
			audiofile.Seek.PosProgres = Scale(audiofile.Seeker.Position(), 0, audiofile.Seeker.Len(), 0, 10000)
			audiofile.Seek.LenProgres = 10000
			audiofile.Seek.PosDisplay = fmt.Sprintf("%02d:%02d\n", int(audiofile.Seek.PosDuration.Minutes())%60, int(audiofile.Seek.PosDuration.Seconds())%60)
			audiofile.Seek.LenDisplay = fmt.Sprintf("%02d:%02d\n", int(audiofile.Seek.LenDuration.Minutes())%60, int(audiofile.Seek.LenDuration.Seconds())%60)
			remaining := audiofile.Seek.LenDuration - audiofile.Seek.PosDuration
			audiofile.Seek.RemainingDisplay = fmt.Sprintf("%02d:%02d\n", int(remaining.Minutes())%60, int(remaining.Seconds())%60)
			status.Id = audiofile.Id
			status.Name = audiofile.Name
			status.Code = audiofile.Code
			status.Seek = audiofile.Seek
			status.Paused = audiofile.Control.Paused
			status.DefaultFadeTime = conf.DefaultFadeTime
			status.Volume = Scale(int(audiofile.Volume.Volume*1000), -6000, 0, 0, 100)
			if remaining == 0 {
				audiofile.Done <- true
			}

			wsret.Play = append(wsret.Play, status)
		} else {
			audiofile.Seek.PosDuration = time.Second * 0
			audiofile.Seek.LenDuration = audiofile.Format.SampleRate.D(audiofile.Buffer.Len())
			audiofile.Seek.PosProgres = 0
			audiofile.Seek.LenProgres = 10000
			audiofile.Seek.PosDisplay = "00:00"

			audiofile.Seek.LenDisplay = fmt.Sprintf("%02d:%02d\n", int(audiofile.Seek.LenDuration.Minutes())%60, int(audiofile.Seek.LenDuration.Seconds())%60)
			remaining := audiofile.Seek.LenDuration - audiofile.Seek.PosDuration
			audiofile.Seek.RemainingDisplay = fmt.Sprintf("%02d:%02d\n", int(remaining.Minutes())%60, int(remaining.Seconds())%60)
			status.Id = audiofile.Id
			status.Name = audiofile.Name
			status.Code = audiofile.Code
			status.Seek = audiofile.Seek
			status.Paused = audiofile.Control.Paused
			status.DefaultFadeTime = conf.DefaultFadeTime
			status.Volume = Scale(int(audiofile.Volume.Volume*1000), -6000, 0, 0, 100)
			wsret.Play = append(wsret.Play, status)
		}
	}

	for _, lib := range AudioFiles {

		wsret.Lib = append(wsret.Lib, WebsocketLibrary{
			Name:            lib.Name,
			Code:            lib.Code,
			DefaultFadeTime: conf.DefaultFadeTime,
		})

	}

	b, err := json.Marshal(wsret)
	if err != nil {
		fmt.Println("json marshal error")
		return
	}
	for conn := range server.clients {
		conn.WriteMessage(websocket.TextMessage, []byte(b))
	}

}
