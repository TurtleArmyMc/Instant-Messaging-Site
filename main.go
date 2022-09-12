package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.StaticFile("/", "static/html/homepage.html")
	r.StaticFile("/javascript/room.js", "static/javascript/room.js")
	r.StaticFile("/css/room.css", "static/css/room.css")

	r.GET("/ping", pong)

	r.Use(setSession)

	r.GET("/rooms/", roomsGET) // Can have roomname in query
	r.GET("/room/:roomname", roomGET)
	r.POST("/room/:roomname", roomPOST)
	r.GET("/room/:roomname/stream", roomStream)
	r.GET("/room/:roomname/history", roomHistory)

	r.Run("0.0.0.0:8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func pong(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func setSession(c *gin.Context) {
	if _, err := c.Cookie("session"); err != nil {
		c.SetCookie("session", newSession(), 60*60*24, "", "", true, true)
	}
}

func newSession() string {
	buff := make([]uint8, 8)
	rand.Read(buff)
	return base64.URLEncoding.EncodeToString(buff)
}

func roomsGET(c *gin.Context) {
	if roomname := c.Request.URL.Query().Get("roomname"); roomname != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/room/"+roomname)
	} else {
		c.HTML(http.StatusOK, "room_index.tmpl.html", gin.H{"Rooms": roomManager.GetRooms()})
	}
}

func roomGET(c *gin.Context) {
	roomName := c.Param("roomname")

	c.HTML(http.StatusOK, "room.tmpl.html", gin.H{
		"RoomName": roomName,
	})
}

func roomPOST(c *gin.Context) {
	roomName := c.Param("roomname")

	var request C2S_CreateMessageRequest
	if err := c.Bind(&request); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	session, err := c.Cookie("session")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	roomManager.PostInRoom(roomName, request.Content, session)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": request.Content,
	})
}

func roomStream(c *gin.Context) {
	roomName := c.Param("roomname")

	roomListener := roomManager.AddUserToRoom(roomName)
	defer roomManager.RemoveUserFromRoom(roomName, roomListener)

	clientGone := c.Writer.CloseNotify()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case message := <-roomListener:
			messageJson, err := message.JSON()
			if err != nil {
				panic(err) // TODO: More graceful error handling
			}
			c.SSEvent("text_message", messageJson)
			return true
		}
	})
}

func roomHistory(c *gin.Context) {
	roomName := c.Param("roomname")

	room := roomManager.getRoom(roomName)
	if room == nil {
		c.JSON(http.StatusNotFound, "No room '"+roomName+"'")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":     room.GetMessages(),
		"endOfHistory": true,
	})
}
