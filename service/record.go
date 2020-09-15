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

type RecordService interface {
	AddRecord(request *entity.Record) error
	FindByFilter(request *entity.Record, page int64, pageSize int64) (records *[]entity.Record, pagination *entity.Pagination, err error)
	Find(page int64, pageSize int64) (records *[]entity.Record, pagination *entity.Pagination, err error)
	ModifyRecord(request *entity.Record) error
	IncreaseFlowerCount(recordId string) (err error)
	IncreaseEggCount(recordId string) (err error)
}

type recordServiceImpl struct {
	client *mongo.Client
}

func init() {
	app.Register(newRecordService)
}

func newRecordService(client *mongo.Client) RecordService {
	return &recordServiceImpl{
		client: client,
	}
}

func (r *recordServiceImpl) AddRecord(request *entity.Record) error {
	if request == nil {
		return errors.New("user is not allowed nil")
	}
	db := r.client.Database("lazybones").Collection("records")
	_, err := db.InsertOne(context.Background(), request)
	return err
}

func (r *recordServiceImpl) FindByFilter(request *entity.Record, page int64, pageSize int64) (records *[]entity.Record, pagination *entity.Pagination, err error) {
	records = &[]entity.Record{}
	returnRecord := entity.Record{}

	db := r.client.Database("lazybones").Collection("records")
	skip := (page - 1) * pageSize
	findOptions := options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
	}

	res, err := db.Find(context.Background(), request, &findOptions)
	for res.Next(context.TODO()) {
		err := res.Decode(&returnRecord)
		if err != nil {
			log.Print(err)
		}
		*records = append(*records, returnRecord)
	}
	count, err := db.CountDocuments(context.Background(), request)
	pagination = &entity.Pagination{
		Page:     int(page),
		PageSize: int(pageSize),
		Count:    int(count),
	}

	return
}

func (r *recordServiceImpl) Find(page int64, pageSize int64) (records *[]entity.Record, pagination *entity.Pagination, err error) {
	records = &[]entity.Record{}

	db := r.client.Database("lazybones").Collection("records")
	skip := (page - 1) * pageSize
	filter := &entity.Record{}

	pipeline := bson.A{
		//多表查询
		bson.D{{"$lookup", bson.D{
			{"from", "users"}, {"localField", "userId"}, {"foreignField", "_id"}, {"as", "userInfo"},
		}}},
		//分页
		bson.D{{"$skip", skip}},
		bson.D{{"$limit", pageSize}},
		//排序
		bson.D{{"$sort", bson.D{{"_id", -1}}}},
	}

	res, err := db.Aggregate(context.Background(), pipeline)

	for res.Next(context.TODO()) {
		//连表查询记录带出其对应用户，不过只需获取到头像和用户名就行
		resultUser := &struct {
			UserInfo []entity.User
		}{}
		returnRecord := entity.Record{}
		//解构出用户信息
		err := res.Decode(&resultUser)
		//解构出记录信息
		err = res.Decode(&returnRecord)
		if err != nil {
			log.Print(err)
		}
		user := resultUser.UserInfo[0]
		returnRecord.UserAvatar = user.Avatar
		returnRecord.UserName = user.UserName
		*records = append(*records, returnRecord)
	}
	count, err := db.CountDocuments(context.Background(), filter)
	pagination = &entity.Pagination{
		Page:     int(page),
		PageSize: int(pageSize),
		Count:    int(count),
	}

	return
}

func (r *recordServiceImpl) ModifyRecord(request *entity.Record) error {
	if request.Id == "" {
		return errors.New("id must not nil")
	}

	id := utils.ToMongoDBId(request.Id)
	//id置为空,配合entity的omitempty查询时忽略空值，不为空的其他字段将更新
	request.Id = ""

	db := r.client.Database("lazybones").Collection("records")
	_, err := db.UpdateOne(context.Background(), bson.D{{"_id", id}}, bson.D{{"$set", request}})
	return err
}

func (r *recordServiceImpl) IncreaseFlowerCount(recordId string) (err error) {
	db := r.client.Database("lazybones").Collection("records")
	filter := bson.D{{"_id", utils.ToMongoDBId(recordId)}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"flower", 1}}}})
	return
}

func (r *recordServiceImpl) IncreaseEggCount(recordId string) (err error) {
	db := r.client.Database("lazybones").Collection("records")
	filter := bson.D{{"_id", utils.ToMongoDBId(recordId)}}
	_, err = db.UpdateOne(context.Background(), filter, bson.D{{"$inc", bson.D{{"egg", 1}}}})
	return
}
