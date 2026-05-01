package models

import "time"

type User struct {
	Id   int       `json:"id"`
	Time time.Time `json:"time`
	Ip   string    `json:"ip"`
}

type Users struct {
	Users []User `json:"users"`
	Count uint64 `json:"count"`
}
