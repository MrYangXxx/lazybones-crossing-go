package entity

import (
	"hidevops.io/hiboot/pkg/model"
	"time"
)

type Record struct {
	model.RequestBody
	Id          string    `json:"_id" bson:"_id,omitempty"`
	Publisher   string    `json:"publisher"` //发布人
	BeginTime   time.Time `json:"beginTime"` //拖延事件开始时间
	EndTime     time.Time `json:"endTime"`   //拖延事件结束时间
	PublishTime time.Time `json:"birthDay"`  //发布世界
	Egg         int       `json:"egg"`       //本记录获得蛋数
	Flower      int       `json:"flower"`    //本记录获得花数
}
