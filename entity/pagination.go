package entity

import "hidevops.io/hiboot/pkg/model"

type Pagination struct {
	model.RequestBody
	Count    int `json:"count"`
	Page     int `json:"page" `
	PageSize int `json:"pageSize" `
}
