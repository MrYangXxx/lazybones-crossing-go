package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"hidevops.io/hiboot-data/starter/mongo"
	"hidevops.io/hiboot/pkg/app"
	"lazybones-crossing-go/entity"
	"lazybones-crossing-go/utils"
	"log"
)

type UserService interface {
	AddUser(user *entity.User) (err error)
	DeleteUser(string) (err error)
	ModifyUser(*entity.User) (err error)
	FindByFilter(interface{}, int64, int64) (users *[]entity.User, pagination *entity.Pagination, err error)
	IsExisted(*entity.User) (existed bool, err error)
	FindById(id string) (user *entity.User, err error)
	IncreaseFlowerCount(userId string) (err error)
	IncreaseEggCount(userId string) (err error)
}

type userServiceImpl struct {
	client *mongo.Client
}

func init() {
	app.Register(newUserService)
}

func newUserService(client *mongo.Client) UserService {
	return &userServiceImpl{
		client: client,
	}
}

func (s *userServiceImpl) AddUser(user *entity.User) (err error) {
	if user == nil {
		return errors.New("user is not allowed nil")
	}
	db := s.client.Database("lazybones").Collection("users")
	_, err = db.InsertOne(context.Background(), user)
	return
}

func (s *userServiceImpl) DeleteUser(id string) (err error) {
	db := s.client.Database("lazybones").Collection("users")
	filter := bson.D{{"id", id}}
	_, err = db.DeleteOne(context.Background(), filter)
	return
}

func (s *userServiceImpl) ModifyUser(user *entity.User) (err error) {
	if user.Id == "" {
		return errors.New("id must not nil")
	}
	db := s.client.Database("lazybones").Collection("users")
	filter := bson.D{{"_id", utils.ToMongoDBId(user.Id)}}
	user.Id = ""
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$set", user}})
	return
}

func (s *userServiceImpl) FindByFilter(filter interface{}, page int64, pageSize int64) (users *[]entity.User, pagination *entity.Pagination, err error) {
	users = &[]entity.User{}

	db := s.client.Database("lazybones").Collection("users")
	skip := (page - 1) * pageSize
	findOptions := options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
	}

	res, err := db.Find(context.Background(), filter, &findOptions)
	for res.Next(context.TODO()) {
		returnUser := entity.User{}
		err := res.Decode(&returnUser)
		if err != nil {
			log.Print(err)
		}
		*users = append(*users, returnUser)
	}
	count, err := db.CountDocuments(context.Background(), filter)
	pagination = &entity.Pagination{
		Page:     int(page),
		PageSize: int(pageSize),
		Count:    int(count),
	}

	return
}

func (s *userServiceImpl) IsExisted(user *entity.User) (existed bool, err error) {

	filter := bson.D{{"$or", bson.D{{"mobile", user.Mobile}, {"email", user.Email}, {"userName", user.UserName}}}}

	db := s.client.Database("lazybones").Collection("users")
	count, err := db.CountDocuments(context.Background(), filter)
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (s *userServiceImpl) FindById(id string) (user *entity.User, err error) {
	if id == "" {
		return nil, errors.New("id is not allowed nil")
	}

	user = &entity.User{}

	db := s.client.Database("lazybones").Collection("users")
	err = db.FindOne(context.Background(), bson.D{{"_id", utils.ToMongoDBId(id)}}).Decode(user)
	return
}

func (s *userServiceImpl) IncreaseFlowerCount(userId string) (err error) {
	db := s.client.Database("lazybones").Collection("users")
	filter := bson.D{{"_id", utils.ToMongoDBId(userId)}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"flower", 1}}}})
	return
}

func (s *userServiceImpl) IncreaseEggCount(userId string) (err error) {
	db := s.client.Database("lazybones").Collection("users")
	filter := bson.D{{"_id", utils.ToMongoDBId(userId)}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"egg", 1}}}})
	return
}
