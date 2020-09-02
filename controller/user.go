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
	"time"
)

// RestController
type userController struct {
	at.RestController
	userService service.UserService
	token       hibootjwt.Token
}

func init() {
	app.Register(newUserController)
}

func newUserController(userService service.UserService, token hibootjwt.Token) *userController {
	return &userController{
		userService: userService,
		token:       token,
	}
}

//密码不能为空，手机号和邮箱其中一个不能为空
func userVerifyEmpty(request *entity.User) error {
	if request.Password == "" || (request.Mobile == "" && request.Email == "") {
		return errors.BadRequestf("必填项不能为空")
	}
	return nil
}

// 用户注册，账号可以为手机或邮箱
func (c *userController) Registry(_ struct {
	at.PostMapping `value:"/registry"`
}, request *entity.User) (model.Response, error) {
	response := new(model.BaseResponse)

	if err := userVerifyEmpty(request); err != nil {
		response.SetCode(http.StatusBadRequest)
		response.SetMessage("必填项不能为空")
		return response, errors.BadRequestf("必填项不能为空")
	}

	//昵称没有设置默认为电话或邮箱
	if request.UserName == "" {
		if request.Mobile == "" {
			request.UserName = request.Email
		}
		request.UserName = request.Mobile
	}

	var page = 1
	var pageSize = 1
	users, _, err := c.userService.FindByFilter(&entity.User{Mobile: request.Mobile, Email: request.Email}, int64(page), int64(pageSize))
	if users != nil && len(*users) > 0 {
		return response, errors.BadRequestf("该账号已存在")
	}

	// MD5 密码盐值
	salt := utils.GetRandomNumber(10)
	request.Password = utils.MD5(salt, request.Password)
	request.Salt = salt

	err = c.userService.AddUser(request)
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

func (c *userController) Login(_ struct {
	at.PostMapping `value:"/login"`
}, request *entity.User) (response model.Response, err error) {
	response = new(model.BaseResponse)
	var page = 1
	var pageSize = 1
	users, _, err := c.userService.FindByFilter(&entity.User{
		Mobile: request.Mobile,
		Email:  request.Email,
	}, int64(page), int64(pageSize))
	if users == nil || len(*users) < 1 {
		return response, errors.BadRequestf("用户不存在")
	}

	password := utils.MD5((*users)[0].Salt, request.Password)
	if (*users)[0].Password != password {
		return response, errors.BadRequestf("用户名或密码错误")
	}

	jwtToken, _ := c.token.Generate(hibootjwt.Map{
		"userName": request.UserName,
		"mobile":   request.Mobile,
		"email":    request.Email,
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
