package entity

import (
	"hidevops.io/hiboot/pkg/model"
	"time"
)

type User struct {
	model.RequestBody
	Id       string    `json:"id" bson:"_id,omitempty"`
	UserName string    `json:"userName" bson:"userName,omitempty"` //昵称
	Password string    `json:"password" bson:"password,omitempty"` //密码
	Email    string    `json:"email" bson:"email,omitempty"`       //邮箱
	Mobile   string    `json:"mobile" bson:"mobile,omitempty"`     //手机
	BirthDay time.Time `json:"birthDay" bson:"birthDay,omitempty"` //生日，暂冗余
	Gender   uint      `json:"gender" bson:"gender,omitempty"`     //性别，暂冗余
	Salt     string    `json:"-" bson:"salt,omitempty"`            //密码加密盐值
	Egg      int       `json:"egg" bson:"egg,omitempty"`           //总获得蛋数
	Flower   int       `json:"flower" bson:"flower,omitempty"`     //总获得花数
	Avatar   string    `json:"avatar" bson:"avatar,omitempty"`     //头像路径
}
