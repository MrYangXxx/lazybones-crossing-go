package entity

import "hidevops.io/hiboot/pkg/model"

type Pagination struct {
	model.RequestBody
	Count    int `json:"count"`     //总记录数
	Page     int `json:"page" `     //页数
	PageSize int `json:"pageSize" ` //没页显示数
}
