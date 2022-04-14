package main

import (
	"fmt"
	"net/http"
	"pixelboard/broker"
	"pixelboard/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	r := gin.Default()

	/*
		GET /canvas キャンバスの状態(すべてのピクセル)を取得する
		POST /canvas/pixels/:y/:x?color=1 キャンバスのピクセルを更新する

		websocket
	*/

	width := 100
	height := 100
	canvas := make([]int, width*height)
	broker := broker.New()

	r.Use(middleware.Cors())

	r.GET("/canvas", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"width":  width,
			"height": height,
			"pixels": canvas,
		})
	})

	r.POST("/canvas/pixels/:y/:x", func(c *gin.Context) {
		y, _ := strconv.Atoi(c.Param("y"))
		x, _ := strconv.Atoi(c.Param("x"))
		color, _ := strconv.Atoi(c.Query("color"))

		if y < 0 || y >= height || x < 0 || x >= width {
			c.JSON(400, gin.H{
				"error": "out of range",
			})
			return
		}

		if color < 0 || color >= 10 {
			c.JSON(400, gin.H{
				"error": "color must be 0-9",
			})
			return
		}

		canvas[y*width+x] = color

		broker.Publish([]int{y, x, color})

		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request, broker)
	})

	r.Run() // listen and serve on⁄
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandler(w http.ResponseWriter, r *http.Request, broker *broker.SimpleBroker) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	sub, err := broker.Subscribe()
	if err != nil {
		fmt.Printf("Failed to subscribe: %+v\n", err)
		return
	}
	defer broker.Unsubscribe(sub)

	for {
		select {
		case msg := <-sub:
			msg, ok := msg.([]int)
			if ok {
				if err := conn.WriteJSON(msg); err != nil {
					fmt.Printf("Failed to send message: %+v\n", err)
					return
				}
			}
		case <-r.Context().Done():
			fmt.Println("Client disconnected")
			return
		}
	}
}
