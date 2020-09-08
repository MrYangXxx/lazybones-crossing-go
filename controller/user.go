package controller

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/juju/errors"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/app/web/context"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	hibootjwt "hidevops.io/hiboot/pkg/starter/jwt"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/service"
	"lazybones-crossing-go/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

// RestController
type userController struct {
	at.RestController
	userService    service.UserService
	captchaService service.CaptchaService
	token          hibootjwt.Token
}

func init() {
	app.Register(newUserController)
}

func newUserController(userService service.UserService, token hibootjwt.Token, captchaService service.CaptchaService) *userController {
	return &userController{
		userService:    userService,
		token:          token,
		captchaService: captchaService,
	}
}

//密码不能为空，手机号和邮箱其中一个不能为空
func userVerifyEmpty(request *entity.User) error {
	if request.Password == "" || (request.Mobile == "" && request.Email == "") {
		return errors.BadRequestf("必填项不能为空")
	}
	return nil
}

//自定义接收参数
type RegistryRequest struct {
	at.RequestBody
	Receiver   string `json:"receiver"`
	Password   string `json:"password"`
	VerifyCode string `json:"verifyCode"`
}

// 用户注册，账号可以为手机或邮箱
func (c *userController) Registry(_ struct {
	at.PostMapping `value:"/registry"`
}, request *RegistryRequest) (model.Response, error) {
	response := new(model.BaseResponse)

	//必填项判空
	hasEmpty := request.Receiver == "" || request.Password == "" || request.VerifyCode == ""
	if hasEmpty {
		response.SetCode(http.StatusBadRequest)
		response.SetMessage("必填项不能为空")
		return response, errors.BadRequestf("必填项不能为空")
	}

	//判断是否是邮箱
	isEmail := utils.VerifyEmailFormat(request.Receiver)
	isMobile := utils.VerifyMobileFormat(request.Receiver)

	if !isEmail && !isMobile {
		return response, errors.BadRequestf("手机或邮箱格式错误")
	}

	//初始化user，先用于查询是否存在，后用于保存
	user := &entity.User{}
	if isEmail {
		user.Email = request.Receiver
	} else {
		user.Mobile = request.Receiver
	}

	//判断用户是否存在
	var page = 1
	var pageSize = 1
	users, _, err := c.userService.FindByFilter(user, int64(page), int64(pageSize))
	if users != nil && len(*users) > 0 {
		return response, errors.BadRequestf("该账号已存在")
	}

	//判断验证码有效性
	captcha := &entity.Captcha{
		Mobile: user.Mobile,
		Email:  user.Email,
		Code:   request.VerifyCode,
		Type:   "registry",
	}
	err = c.captchaService.FindCaptcha(captcha)
	// EqualFold方法可忽略大小写，相等返回true
	if err != nil || !strings.EqualFold(captcha.Code, request.VerifyCode) {
		log.Print("验证码验证失败")
		return response, errors.BadRequestf("验证码验证失败")
	}
	if time.Now().After(captcha.Expired) {
		log.Print("验证码过期")
		return response, errors.BadRequestf("验证码已过期")
	}

	// MD5 密码盐值
	salt := utils.GetRandomNumber(10)
	//设置要保存的user信息
	user.Password = utils.MD5(salt, request.Password)
	user.Salt = salt
	//昵称没有设置默认为电话或邮箱
	user.UserName = request.Receiver

	err = c.userService.AddUser(user)
	response.SetData(request)
	return response, err
}

//PUT /user/id/:id
func (c *userController) PutById(id string, request *entity.User) (response model.Response, err error) {
	response = new(model.BaseResponse)
	request.Id = id

	if err := userVerifyEmpty(request); err != nil {
		response.SetCode(http.StatusBadRequest)
		response.SetMessage("必填项不能为空")
		return response, errors.BadRequestf("必填项不能为空")
	}

	users, _, err := c.userService.FindByFilter(&entity.User{Id: id}, 1, 1)
	if users == nil || len(*users) < 1 {
		return response, errors.BadRequestf("该账号不存在")
	}
	//密码被改了
	if (*users)[0].Password != request.Password {
		request.Password = utils.MD5((*users)[0].Salt, request.Password)
	}
	request.Salt = (*users)[0].Salt

	err = c.userService.ModifyUser(request)

	return
}

//自定义接收参数
type LoginRequest struct {
	at.RequestBody
	Receiver string `json:"receiver"`
	Password string `json:"password"`
}

func (c *userController) Login(_ struct {
	at.PostMapping `value:"/login"`
}, request *LoginRequest) (response model.Response, err error) {
	response = new(model.BaseResponse)

	filter := &entity.User{}
	//判断是否是邮箱
	isEmail := utils.VerifyEmailFormat(request.Receiver)

	if isEmail {
		filter.Email = request.Receiver
	} else {
		filter.Mobile = request.Receiver
	}

	var page = 1
	var pageSize = 1
	users, _, err := c.userService.FindByFilter(filter, int64(page), int64(pageSize))
	if users == nil || len(*users) < 1 {
		return response, errors.BadRequestf("用户不存在")
	}

	password := utils.MD5((*users)[0].Salt, request.Password)
	if (*users)[0].Password != password {
		return response, errors.BadRequestf("用户名或密码错误")
	}

	jwtToken, _ := c.token.Generate(hibootjwt.Map{
		"userName": (*users)[0].UserName,
		"mobile":   (*users)[0].Mobile,
		"email":    (*users)[0].Email,
	}, 10, time.Hour)

	data := make(map[string]interface{})
	data["token"] = jwtToken

	response.SetData(data)
	return
}

func (c *userController) Info(_ struct {
	at.GetMapping `value:"/info"`
}, ctx context.Context) (response model.Response, err error) {
	response = new(model.BaseResponse)
	data := make(map[string]interface{})

	//获取token
	token, _ := request.ParseFromRequest(ctx.Request(), request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		return c.token.VerifyKey(), nil
	})

	//获取用户信息
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		users, _, err := c.userService.FindByFilter(&entity.User{
			Mobile: claims["mobile"].(string),
			Email:  claims["email"].(string),
		}, 1, 1)
		if err != nil {
			log.Print(err)
		}
		data["userInfo"] = (*users)[0]
	}

	response.SetData(data)
	return
}
