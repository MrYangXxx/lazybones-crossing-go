package entity

//用户给记录投鲜花鸡蛋实体
type UserRecord struct {
	UserId   string `json:"userId" bson:"userId,omitempty"`
	RecordId string `json:"recordId" bson:"recordId,omitempty"`
	Egg      int    `json:"egg"`    //用户投鸡蛋数，暂定限量10
	Flower   int    `json:"flower"` //用户投鲜花数，暂定限量10
}
