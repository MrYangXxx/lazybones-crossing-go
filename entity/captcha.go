package entity

import (
	"hidevops.io/hiboot/pkg/model"
	"time"
)

//验证码
type Captcha struct {
	model.RequestBody
	Id      string    `json:"_id" bson:"_id,omitempty"`
	Mobile  string    `json:"mobile"`  //发送的手机号
	Email   string    `json:"email"`   //发送的邮箱
	Code    string    `json:"code"`    //验证码
	Type    string    `json:"type"`    //验证码用途
	Expired time.Time `json:"expired"` //过期时间
}
