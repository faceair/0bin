package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const letterBytes = "01234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

type Data map[string]*struct {
	Text     []byte
	CreateAt time.Time
}

func (d Data) Add(name string, text []byte) {
	d[name] = &struct {
		Text     []byte
		CreateAt time.Time
	}{
		Text:     text,
		CreateAt: time.Now(),
	}
}

func (d Data) Get(name string) []byte {
	if value, ok := d[name]; ok {
		defer delete(d, name)
		if time.Now().Sub(value.CreateAt) < time.Hour*24 {
			return value.Text
		}
	}
	return []byte{}
}

func (d Data) Start() {
	go func() {
		for {
			for name, value := range d {
				if time.Now().Sub(value.CreateAt) > time.Hour*24 {
					delete(d, name)
				}
			}
			time.Sleep(time.Minute)
		}
	}()
}

func main() {
	index, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}

	data := Data{}
	data.Start()

	panic(http.ListenAndServe(":8061", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/"+randString(5)+"#new", 302)
		}
		if strings.HasPrefix(r.URL.Path, "/api") {
			if r.Method == "GET" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write(data.Get(r.URL.Path))
				return
			} else if r.Method == "POST" {
				text, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "read body failed", 400)
					return
				}
				data.Add(r.URL.Path, text)
				w.Write([]byte{})
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})))
}
