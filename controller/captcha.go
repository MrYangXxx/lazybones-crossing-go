package controller

import (
	"github.com/juju/errors"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/service"
	"lazybones-crossing-go/utils"
	"time"
)

// RestController
type captchaController struct {
	at.RestController
	captchaService service.CaptchaService
}

func init() {
	app.Register(newCaptchaController)
}

func newCaptchaController(captchaService service.CaptchaService) *captchaController {
	return &captchaController{
		captchaService: captchaService,
	}
}

// 发送验证码
func (c *captchaController) Post(request *entity.Captcha) (model.Response, error) {
	response := new(model.BaseResponse)

	canSendCaptcha := request.Type != "" && (request.Email != "" || request.Mobile != "")
	if !canSendCaptcha {
		return response, errors.BadRequestf("请填写账号")
	}

	code := utils.GetRandomString(5)
	c.captchaService.AddCaptcha(&entity.Captcha{
		RequestBody: model.RequestBody{},
		Mobile:      request.Mobile,
		Email:       request.Email,
		Code:        code,
		Type:        request.Type,
		Expired:     time.Now().Add(time.Minute * 15).Local(),
	})

	response.SetData(code)
	return response, nil
}

//测试验证码查找
// 发送验证码
func (c *captchaController) Find(_ struct {
	at.PostMapping `value:"/find"`
}, request *entity.Captcha) (model.Response, error) {
	response := new(model.BaseResponse)
	captcha := &entity.Captcha{
		Mobile: request.Mobile,
		Email:  request.Email,
		Type:   request.Type,
	}

	err := c.captchaService.FindCaptchaByMobileOrEmail(captcha)

	if err != nil {
		return response, err
	}

	response.SetData(captcha)
	return response, nil
}
