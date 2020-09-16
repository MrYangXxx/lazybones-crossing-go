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

var FLOWERLOCK int32 = 1

func (c *recordController) Flower(_ struct {
	at.PostMapping `value:"/flower"`
}, request *entity.Record) (response model.Response, err error) {
	response = new(model.BaseResponse)
	ok := atomic.CompareAndSwapInt32(&FLOWERLOCK, 1, 0)
	defer atomic.CompareAndSwapInt32(&FLOWERLOCK, 0, 1)
	if !ok {
		return response, errors.BadRequestf("操作过快，请重试")
	}

	//查询用户投票记录需要
	hasEmpty := request.Id == "" || request.UserId == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}

	//查询用户投票记录
	userRecord, err := c.userRecordService.Find(request.UserId.(string), request.Id)
	if err != nil || userRecord == nil {
		userRecord = &entity.UserRecord{
			UserId:   request.UserId.(string),
			RecordId: request.Id,
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
	//用户总鲜花数+1
	err = c.userService.IncreaseFlowerCount(request.UserId.(string))
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}
	//记录鲜花数+1
	err = c.recordService.IncreaseFlowerCount(request.Id)
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}
	//用户投送鲜花数+1
	err = c.userRecordService.IncreaseFlowerCount(request.UserId.(string), request.Id)
	if err != nil {
		return response, errors.BadRequestf("投递鲜花失败")
	}

	//返回前端
	request.Flower += 1
	response.SetData(request)

	return
}

var EGGLOCK int32 = 1

func (c *recordController) EGG(_ struct {
	at.PostMapping `value:"/egg"`
}, request *entity.Record) (response model.Response, err error) {
	response = new(model.BaseResponse)
	ok := atomic.CompareAndSwapInt32(&EGGLOCK, 1, 0)
	defer atomic.CompareAndSwapInt32(&EGGLOCK, 0, 1)
	if !ok {
		return response, errors.BadRequestf("操作过快，请重试")
	}

	//查询用户投票记录需要
	hasEmpty := request.Id == "" || request.UserId == ""

	if hasEmpty {
		return response, errors.BadRequestf("传输数据不完整")
	}

	//查询用户投票记录
	userRecord, err := c.userRecordService.Find(request.UserId.(string), request.Id)
	if err != nil || userRecord == nil {
		userRecord = &entity.UserRecord{
			UserId:   request.UserId.(string),
			RecordId: request.Id,
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
	//用户总蛋数+1
	err = c.userService.IncreaseEggCount(request.UserId.(string))
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}
	//记录蛋数+1
	err = c.recordService.IncreaseEggCount(request.Id)
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}
	//用户投送蛋数+1
	err = c.userRecordService.IncreaseEggCount(request.UserId.(string), request.Id)
	if err != nil {
		return response, errors.BadRequestf("投递鸡蛋失败")
	}

	//返回前端
	request.Egg += 1
	response.SetData(request)
	return
}
