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

//自定义接收参数
type CaptchaRequest struct {
	at.RequestBody
	Receiver string `json:"receiver"`
	Type     string `json:"type"`
}

// 发送验证码
func (c *captchaController) Post(request *CaptchaRequest) (model.Response, error) {
	response := new(model.BaseResponse)

	canSendCaptcha := request.Type != "" && request.Receiver != ""
	if !canSendCaptcha {
		return response, errors.BadRequestf("请填写账号")
	}

	//判断是否是邮箱
	isEmail := utils.VerifyEmailFormat(request.Receiver)
	isMobile := utils.VerifyMobileFormat(request.Receiver)

	if !isEmail && !isMobile {
		return response, errors.BadRequestf("手机或邮箱格式错误")
	}

	//生成5位随机字符串
	code := utils.GetRandomString(5)
	//验证码实体初始化，过期时间设置为15分钟
	captcha := &entity.Captcha{
		RequestBody: model.RequestBody{},
		Code:        code,
		Type:        request.Type,
		Expired:     time.Now().Add(time.Minute * 15).Local(),
	}

	if isEmail {
		captcha.Email = request.Receiver
	} else {
		captcha.Mobile = request.Receiver
	}

	err := c.captchaService.AddCaptcha(captcha)

	if err != nil {
		log.Print(err)
		return nil, errors.BadRequestf("发送验证码失败")
	}

	if captcha.Email != "" {
		err = utils.SendMail([]string{captcha.Email}, "[集合吧懒虫们]验证码", "你正在进行登录操作，你的验证码为:"+code)
		if err != nil {
			log.Print("发送邮件验证码失败")
		}
	}

	if captcha.Mobile != "" {
		//todo 发送手机验证码
	}

	log.Print("验证码为:" + code)

	return response, nil
}

//测试验证码查找
func (c *captchaController) Find(_ struct {
	at.PostMapping `value:"/find"`
}, request *entity.Captcha) (model.Response, error) {
	response := new(model.BaseResponse)
	captcha := &entity.Captcha{
		Mobile: request.Mobile,
		Email:  request.Email,
		Type:   request.Type,
	}

	err := c.captchaService.FindCaptcha(captcha)

	if err != nil {
		return response, err
	}

	response.SetData(captcha)
	return response, nil
}
