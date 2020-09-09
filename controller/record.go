package controller

import (
	"github.com/juju/errors"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/service"
	"log"
	"time"
)

// RestController
type recordController struct {
	at.RestController
	recordService service.RecordService
}

func init() {
	app.Register(newRecordController)
}

func newRecordController(recordService service.RecordService) *recordController {
	return &recordController{
		recordService: recordService,
	}
}

//发布拖延记录
func (c *recordController) Post(request *entity.Record) (model.Response, error) {
	response := new(model.BaseResponse)

	hasEmpty := request.UserId == "" || request.BeginTime.String() == "" || request.EndTime.String() == ""
	if hasEmpty {
		return response, errors.BadRequestf("请填写账号")
	}

	record := &entity.Record{
		UserId:      request.UserId,
		BeginTime:   request.BeginTime,
		EndTime:     request.EndTime,
		PublishTime: time.Now(),
		Content:     request.Content,
		Egg:         0,
		Flower:      0,
		Complete:    false,
	}

	c.recordService.AddRecord(record)

	response.SetData(record)
	return response, nil
}

//查询某用户发布记录
func (c *recordController) FindByUserId(_ struct {
	at.PostMapping `value:"/user"`
}, request *struct {
	at.RequestBody
	UserId   string
	Page     int
	PageSize int
}) (model.Response, error) {
	response := new(model.BaseResponse)

	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = 10
	}

	records, pagination, err := c.recordService.FindByFilter(&entity.Record{UserId: request.UserId}, int64(request.Page), int64(request.PageSize))
	if err != nil {
		log.Print(err)
		return nil, errors.BadRequestf("查询用户发布记录失败")
	}

	data := make(map[string]interface{})
	data["records"] = records
	data["pagination"] = pagination
	response.SetData(data)
	return response, nil
}

//首页显示的发布记录,暂根据发布时间排序查询
func (c *recordController) Find(_ struct {
	at.PostMapping `value:"/find"`
}, request *struct {
	at.RequestBody
	Page     int
	PageSize int
}) (model.Response, error) {
	response := new(model.BaseResponse)

	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = 10
	}
	records, pagination, err := c.recordService.Find(int64(request.Page), int64(request.PageSize))
	if err != nil {
		log.Print(err)
		return nil, errors.BadRequestf("查询发布记录失败")
	}

	data := make(map[string]interface{})
	data["records"] = records
	data["pagination"] = pagination
	response.SetData(data)
	return response, nil
}

func (c *recordController) Modify(_ struct {
	at.PostMapping `value:"/modify"`
}, request *entity.Record) (response model.Response, err error) {
	response = new(model.BaseResponse)

	hasEmpty := request.Id == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}

	err = c.recordService.ModifyRecord(request)
	return
}
