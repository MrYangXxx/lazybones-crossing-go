package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"hidevops.io/hiboot-data/starter/mongo"
	"hidevops.io/hiboot/pkg/app"
	"lazybones-crossing-go/entity"
	"log"
)

type CaptchaService interface {
	AddCaptcha(captcha *entity.Captcha) error
	FindCaptcha(captcha *entity.Captcha) error
}

type captchaServiceImpl struct {
	client *mongo.Client
}

func init() {
	app.Register(newCaptchaService)
}

func newCaptchaService(client *mongo.Client) CaptchaService {
	return &captchaServiceImpl{
		client: client,
	}
}

func (c *captchaServiceImpl) AddCaptcha(captcha *entity.Captcha) (err error) {
	if captcha == nil {
		return errors.New("user is not allowed nil")
	}
	db := c.client.Database("lazybones").Collection("captchas")
	_, err = db.InsertOne(context.Background(), captcha)
	return
}

//根据email或mobile，并且等于type 查询
func (c *captchaServiceImpl) FindCaptcha(captcha *entity.Captcha) error {
	if captcha == nil {
		return errors.New("captcha is not allowed nil")
	}
	filter := bson.D{{"type", captcha.Type}, {"$or", bson.A{bson.D{{"mobile", captcha.Mobile}}, bson.D{{"email", captcha.Email}}}}}
	findOption := options.FindOne().SetSort(bson.D{{"_id", -1}})

	db := c.client.Database("lazybones").Collection("captchas")
	err := db.FindOne(context.Background(), filter, findOption).Decode(&captcha)
	if err != nil {
		log.Print(err)
	}
	return nil
}
