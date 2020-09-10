package entity

import (
	"hidevops.io/hiboot/pkg/model"
	"time"
)

type Record struct {
	model.RequestBody
	Id          string    `json:"id" bson:"_id,omitempty"`
	UserId      string    `json:"userId" bson:"userId,omitempty"`           //发布人Id
	UserName    string    `json:"userName" bson:"-"`                        //发布人,-为不保存于数据库
	UserAvatar  string    `json:"userAvatar" bson:"-"`                      //头像路径
	BeginTime   time.Time `json:"beginTime" bson:"beginTime,omitempty"`     //拖延事件开始时间
	EndTime     time.Time `json:"endTime" bson:"endTime,omitempty"`         //拖延事件结束时间
	Content     string    `json:"content" bson:"content,omitempty"`         //发布内容
	PublishTime time.Time `json:"publishTime" bson:"publishTime,omitempty"` //发布时间
	Egg         int       `json:"egg" bson:"egg,omitempty"`                 //本记录获得蛋数(观众认为完成不了)
	Flower      int       `json:"flower" bson:"flower,omitempty"`           //本记录获得花数(观众认为能够完成)
	Complete    bool      `json:"complete"`                                 //结束时间到达后此拖延事件是否完成
}
