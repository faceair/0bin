package main

import "time"

type DataBase map[string]*struct {
	Text     []byte
	CreateAt time.Time
}

func (d DataBase) Add(name string, text []byte) {
	d[name] = &struct {
		Text     []byte
		CreateAt time.Time
	}{
		Text:     text,
		CreateAt: time.Now(),
	}
}

func (d DataBase) Get(name string) []byte {
	if value, ok := d[name]; ok {
		defer delete(d, name)
		if time.Now().Sub(value.CreateAt) < time.Hour*24 {
			return value.Text
		}
	}
	return nil
}

func (d DataBase) Start() {
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
