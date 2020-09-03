package entity

import (
	"hidevops.io/hiboot/pkg/model"
	"time"
)

type User struct {
	model.RequestBody
	Id       string    `json:"_id" bson:"_id,omitempty"`
	UserName string    `json:"userName" bson:"userName,omitempty"` //昵称
	Password string    `json:"password" bson:"password,omitempty"`
	Email    string    `json:"email" bson:"email,omitempty"`
	Mobile   string    `json:"mobile" bson:"mobile,omitempty"`
	BirthDay time.Time `json:"birthDay" bson:"birthDay,omitempty"`
	Gender   uint      `json:"gender" bson:"gender,omitempty"`
	Salt     string    `json:"-" bson:"salt,omitempty"`
	Egg      int       `json:"egg" bson:"egg,omitempty"`       //总获得蛋数
	Flower   int       `json:"flower" bson:"flower,omitempty"` //总获得花数
}
