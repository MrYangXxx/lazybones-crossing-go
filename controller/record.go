package controller

import (
	"github.com/juju/errors"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/service"
	"lazybones-crossing-go/utils"
	"log"
	"sync/atomic"
	"time"
)

// RestController
type recordController struct {
	at.RestController
	recordService     service.RecordService
	userRecordService service.UserRecordService
	userService       service.UserService
}

func init() {
	app.Register(newRecordController)
}

func newRecordController(recordService service.RecordService, userRecordService service.UserRecordService, userService service.UserService) *recordController {
	return &recordController{
		recordService:     recordService,
		userRecordService: userRecordService,
		userService:       userService,
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
		UserId:      utils.ToMongoDBId(request.UserId.(string)),
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

	filter := make(map[string]interface{})
	filter["userId"] = utils.ToMongoDBId(request.UserId)

	records, pagination, err := c.recordService.FindByFilter(filter, int64(request.Page), int64(request.PageSize))
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

//记录修改，只能修改完成状态
func (c *recordController) Modify(_ struct {
	at.PostMapping `value:"/modify"`
}, request *entity.Record) (response model.Response, err error) {
	response = new(model.BaseResponse)

	hasEmpty := request.Id == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}
	update := make(map[string]interface{})
	update["complete"] = true

	err = c.recordService.ModifyRecord(request.Id, update)
	return
}

//投送鲜花鸡蛋自定义接收参数
type FlowerAndEggRequest struct {
	at.RequestBody
	RecordId string `json:"recordId"` //记录id
	UserId   string `json:"userId"`   //被送人id
	OwnerId  string `json:"ownerId"`  //赠送人(自己)id
}

var FLOWERLOCK int32 = 1

func (c *recordController) Flower(_ struct {
	at.PostMapping `value:"/flower"`
}, request *FlowerAndEggRequest) (response model.Response, err error) {
	response = new(model.BaseResponse)
	ok := atomic.CompareAndSwapInt32(&FLOWERLOCK, 1, 0)
	defer atomic.CompareAndSwapInt32(&FLOWERLOCK, 0, 1)
	if !ok {
		return response, errors.BadRequestf("操作过快，请重试")
	}

	//查询用户投票记录需要
	hasEmpty := request.OwnerId == "" || request.UserId == "" || request.RecordId == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}

	//查询用户（本人）投票记录
	userRecord, err := c.userRecordService.Find(request.OwnerId, request.RecordId)
	if err != nil || userRecord == nil {
		userRecord = &entity.UserRecord{
			UserId:   request.OwnerId,
			RecordId: request.RecordId,
			Egg:      0,
			Flower:   0,
		}
		err = c.userRecordService.Add(userRecord)
		if err != nil {
			return response, errors.BadRequestf("投递鲜花失败")
		}
	}

	//用户对记录的投花上限暂定为10
	if userRecord.Flower >= 10 {
		return response, errors.BadRequestf("您在此记录的投花数已达上限")
	}

	//投送数量未超过时
	//用户(被赠送人)总鲜花数+1
	err = c.userService.IncreaseFlowerCount(request.UserId)
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}
	//记录鲜花数+1
	err = c.recordService.IncreaseFlowerCount(request.RecordId)
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}
	//用户(本人)投送鲜花数+1
	err = c.userRecordService.IncreaseFlowerCount(request.OwnerId, request.RecordId)
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}

	return
}

var EGGLOCK int32 = 1

func (c *recordController) EGG(_ struct {
	at.PostMapping `value:"/egg"`
}, request *FlowerAndEggRequest) (response model.Response, err error) {
	response = new(model.BaseResponse)
	ok := atomic.CompareAndSwapInt32(&EGGLOCK, 1, 0)
	defer atomic.CompareAndSwapInt32(&EGGLOCK, 0, 1)
	if !ok {
		return response, errors.BadRequestf("操作过快，请重试")
	}

	//查询用户投票记录需要
	hasEmpty := request.OwnerId == "" || request.UserId == "" || request.RecordId == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}

	//查询用户(本人)投票记录
	userRecord, err := c.userRecordService.Find(request.OwnerId, request.RecordId)
	if err != nil || userRecord == nil {
		userRecord = &entity.UserRecord{
			UserId:   request.OwnerId,
			RecordId: request.RecordId,
			Egg:      0,
			Flower:   0,
		}
		err = c.userRecordService.Add(userRecord)
		if err != nil {
			return response, errors.BadRequestf("投递鸡蛋失败")
		}
	}

	//用户对记录的投蛋上限暂定为10
	if userRecord.Egg >= 10 {
		return response, errors.BadRequestf("您在此记录的投蛋数已达上限")
	}

	//投送数量未超过时
	//用户（被赠送人）总蛋数+1
	err = c.userService.IncreaseEggCount(request.UserId)
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}
	//记录蛋数+1
	err = c.recordService.IncreaseEggCount(request.RecordId)
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}
	//用户（本人）投送蛋数+1
	err = c.userRecordService.IncreaseEggCount(request.OwnerId, request.RecordId)
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}

	return
}

//记录删除
func (c *recordController) Delete(_ struct {
	at.PostMapping `value:"/delete"`
}, request *entity.Record) (response model.Response, err error) {
	response = new(model.BaseResponse)
	err = c.recordService.DeleteRecord(request.Id)
	if err != nil {
		return response, errors.BadRequestf("删除记录失败")
	}
	return
}
