package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"hidevops.io/hiboot-data/starter/mongo"
	"hidevops.io/hiboot/pkg/app"
	"lazybones-crossing-go/entity"
)

type UserRecordService interface {
	Find(userId, recordId string) (userRecord *entity.UserRecord, err error)
	Add(request *entity.UserRecord) (err error)
	IncreaseFlowerCount(userId, recordId string) (err error)
	IncreaseEggCount(userId, recordId string) (err error)
}

type userRecordServiceImpl struct {
	// add UserService, it means that the instance of UserServiceImpl can be found by UserService
	client *mongo.Client
}

func init() {
	app.Register(newUserRecordService)
}

func newUserRecordService(client *mongo.Client) UserRecordService {
	return &userRecordServiceImpl{
		client: client,
	}
}

func (u userRecordServiceImpl) Add(request *entity.UserRecord) (err error) {
	if request == nil {
		return errors.New("user is not allowed nil")
	}
	db := u.client.Database("lazybones").Collection("user_record")
	_, err = db.InsertOne(context.Background(), request)
	return
}

func (u userRecordServiceImpl) IncreaseFlowerCount(userId, recordId string) (err error) {
	db := u.client.Database("lazybones").Collection("user_record")
	filter := bson.D{{"userId", userId}, {"recordId", recordId}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"flower", 1}}}})
	return
}

func (u userRecordServiceImpl) IncreaseEggCount(userId, recordId string) (err error) {
	db := u.client.Database("lazybones").Collection("user_record")
	filter := bson.D{{"userId", userId}, {"recordId", recordId}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"egg", 1}}}})
	return
}

func (u userRecordServiceImpl) Find(userId, recordId string) (userRecord *entity.UserRecord, err error) {
	db := u.client.Database("lazybones").Collection("user_record")
	filter := bson.D{{"userId", userId}, {"recordId", recordId}}

	userRecord = &entity.UserRecord{}
	err = db.FindOne(context.Background(), filter).Decode(userRecord)
	return
}
