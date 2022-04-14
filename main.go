package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pixelboard/broker"
	"pixelboard/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	r := gin.Default()

	width := 100
	height := 100
	canvas := make([]int, width*height)
	isDirty := false
	broker := broker.New()

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(os.Getenv("MONGODB_URI")).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("pxlbrd").Collection("canvas")

	// canvasをロード
	if res := collection.FindOne(ctx, bson.D{}); res.Err() == nil {
		var result struct {
			Canvas []int
		}
		err = res.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		canvas = result.Canvas
	}

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
		isDirty = true

		broker.Publish([]int{y, x, color})

		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request, broker)
	})

	go func() {
		for {
			if isDirty {
				isDirty = false

				ctx := context.Background()
				_, err := collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"canvas", canvas}}}}, options.Update().SetUpsert(true))
				if err != nil {
					log.Println(err)
				}
			}
			time.Sleep(time.Second)
		}
	}()

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
