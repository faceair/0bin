package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
)

var indexFile []byte

var DB DataBase
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	var err error
	indexFile, err = ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}

	DB = DataBase{}
	DB.Start()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer conn.Close()
	path := r.URL.Query().Get("path")
	err = conn.WriteMessage(1, DB.Get(path))
	if err != nil {
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		DB.Add(path, message)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/"+randString(5), 302)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexFile)
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", httpHandler)
	panic(http.ListenAndServe(":8061", nil))
}
