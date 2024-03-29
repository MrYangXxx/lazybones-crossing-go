package controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/juju/errors"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/app/web/context"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	hibootjwt "hidevops.io/hiboot/pkg/starter/jwt"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/middleware"
	"lazybones-crossing-go/service"
	"lazybones-crossing-go/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// RestController
type userController struct {
	at.RestController
	userService    service.UserService
	captchaService service.CaptchaService
	token          hibootjwt.Token
	file           *middleware.File
}

func init() {
	app.Register(newUserController)
}

func newUserController(userService service.UserService, token hibootjwt.Token, captchaService service.CaptchaService, file *middleware.File) *userController {
	return &userController{
		userService:    userService,
		token:          token,
		captchaService: captchaService,
		file:           file,
	}
}

//密码不能为空，手机号和邮箱其中一个不能为空
func userVerifyEmpty(request *entity.User) error {
	if request.Password == "" || (request.Mobile == "" && request.Email == "") {
		return errors.BadRequestf("必填项不能为空")
	}
	return nil
}

//注册自定义接收参数
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
	filter := make(map[string]string)
	user := &entity.User{}
	if isEmail {
		filter["email"] = request.Receiver
		user.Email = request.Receiver
	} else {
		filter["mobile"] = request.Receiver
		user.Mobile = request.Receiver
	}

	//判断用户是否存在
	var page = 1
	var pageSize = 1
	users, _, err := c.userService.FindByFilter(filter, int64(page), int64(pageSize))
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
	//默认头像
	user.Avatar = "default.jpg"

	err = c.userService.AddUser(user)
	response.SetData(request)
	return response, err
}

//PUT /user/id/:id 用户修改
func (c *userController) PutById(id string, request *entity.User) (response model.Response, err error) {
	response = new(model.BaseResponse)

	if err := userVerifyEmpty(request); err != nil {
		response.SetCode(http.StatusBadRequest)
		response.SetMessage("必填项不能为空")
		return response, errors.BadRequestf("必填项不能为空")
	}

	user, err := c.userService.FindById(id)
	if user == nil {
		return response, errors.BadRequestf("该账号不存在")
	}
	//密码被改了
	if user.Password != request.Password {
		request.Password = utils.MD5(user.Salt, request.Password)
	}
	request.Salt = user.Salt

	//头像被改了，删除旧的头像
	if user.Avatar != request.Avatar {
		//存储路径
		path := c.file.Path
		err = os.Remove(path + user.Avatar)
		if err != nil {
			fmt.Println("删除头像文件出现错误", err)
		}
	}

	err = c.userService.ModifyUser(request)

	return
}

//登录自定义接收参数
type LoginRequest struct {
	at.RequestBody
	Receiver string `json:"receiver"`
	Password string `json:"password"`
}

//用户登录
func (c *userController) Login(_ struct {
	at.PostMapping `value:"/login"`
}, request *LoginRequest) (response model.Response, err error) {
	response = new(model.BaseResponse)

	filter := make(map[string]interface{})
	//判断是否是邮箱
	isEmail := utils.VerifyEmailFormat(request.Receiver)

	if isEmail {
		filter["email"] = request.Receiver
	} else {
		filter["mobile"] = request.Receiver
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

//用户信息
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
		filter := make(map[string]interface{})
		if claims["mobile"] != "" {
			filter["mobile"] = claims["mobile"]
		}
		if claims["email"] != "" {
			filter["email"] = claims["email"]
		}
		users, _, err := c.userService.FindByFilter(filter, 1, 1)
		if err != nil {
			log.Print(err)
		}
		data["userInfo"] = (*users)[0]
	}

	response.SetData(data)
	return
}

//查询用户排行榜
func (c *userController) Leaderboard(_ struct {
	at.PostMapping `value:"/leaderboard"`
}, request *struct {
	at.RequestBody
	Sort     string
	Page     int
	PageSize int
}) (response model.Response, err error) {
	response = new(model.BaseResponse)
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = 10
	}
	users, pagination, err := c.userService.FindBySort(request.Sort, int64(request.Page), int64(request.PageSize))
	if users == nil || len(*users) < 1 {
		return response, errors.BadRequestf("用户排行榜查询失败")
	}
	data := make(map[string]interface{})
	data["users"] = users
	data["pagination"] = pagination
	response.SetData(data)
	return
}
