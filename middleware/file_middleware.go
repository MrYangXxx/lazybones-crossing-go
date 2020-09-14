package middleware

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/app/web/context"
	"hidevops.io/hiboot/pkg/at"
	"hidevops.io/hiboot/pkg/model"
	hibootjwt "hidevops.io/hiboot/pkg/starter/jwt"
	"io"
	"io/ioutil"
	"lazybones-crossing-go/utils"
	"log"
	"os"
)

type File struct {
	at.ConfigurationProperties `value:"file"`
	Path                       string `json:"path"`
}

type fileMiddleware struct {
	at.RestController
	file  *File
	token hibootjwt.Token
}

func init() {
	app.Register(newFileMiddleware, new(File))
}

func newFileMiddleware(file *File, token hibootjwt.Token) *fileMiddleware {
	return &fileMiddleware{file: file, token: token}
}

func (c *fileMiddleware) Upload(_ struct {
	at.PostMapping `value:"/upload"`
}, ctx context.Context) (response model.Response, err error) {
	response = new(model.BaseResponse)
	ctx.Request().ParseMultipartForm(32 << 20)
	file, _, err := ctx.Request().FormFile("file")
	if err != nil {
		log.Println("form file err: ", err)
		return
	}
	defer file.Close()

	fileName := ""

	//获取token
	token, _ := jwt.Parse(ctx.GetCookie("lazybones_token"), func(*jwt.Token) (interface{}, error) {
		return c.token.VerifyKey(), nil
	})

	//获取用户信息
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		mobile := claims["mobile"].(string)
		email := claims["email"].(string)
		if mobile != "" {
			fileName = mobile + utils.GetRandomNumber(5) + ".jpg"
		} else {
			fileName = email + utils.GetRandomNumber(5) + ".jpg"
		}

	}

	//选择存储路径
	path := c.file.Path

	//判断存储路径是否存在
	_, err = os.Stat(path)
	if err != nil {
		os.Mkdir(path, os.ModePerm)
	}

	//创建上传的目的文件
	f, err := os.OpenFile(path+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("open file err: ", err)
		return
	}
	defer f.Close()
	//拷贝文件
	_, err = io.Copy(f, file)
	if err != nil {
		log.Println("copy file err: ", err)
		return
	}

	data := make(map[string]interface{})
	data["fileName"] = fileName

	response.SetData(data)

	return
}

func (c *fileMiddleware) Remove(_ struct {
	at.PostMapping `value:"/remove"`
}, request *struct {
	model.RequestBody
	FileName string
}) (response model.Response, err error) {
	response = new(model.BaseResponse)

	//选择存储路径
	path := c.file.Path

	err = os.Remove(path + request.FileName)

	if err != nil {
		fmt.Println("open file err: ", err)
		return response, errors.New("删除出现错误")
	}

	return response, nil
}

func (c *fileMiddleware) DownLoad(_ struct {
	at.GetMapping `value:"/download/{fileName}"`
}, fileName string, ctx context.Context) {
	//选择存储路径
	path := c.file.Path

	rw := ctx.ResponseWriter()
	file, err := os.Open(path + fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()
	rw.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	rw.Header().Set("Content-Type", ctx.Request().Header.Get("Content-Type"))
	io.Copy(rw, file)
}

func (c *fileMiddleware) Image(_ struct {
	at.GetMapping `value:"/img/{fileName}"`
}, fileName string, ctx context.Context) {
	//选择存储路径
	path := c.file.Path
	rw := ctx.ResponseWriter()
	file, err := os.Open(path + fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()
	buff, err := ioutil.ReadAll(file)
	rw.Write(buff)
}
