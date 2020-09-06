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
	returnRecord := entity.Record{}

	db := r.client.Database("lazybones").Collection("records")
	skip := (page - 1) * pageSize
	findOptions := &options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
		Sort:  bson.D{{"_id", -1}},
	}

	filter := &entity.Record{}

	res, err := db.Find(context.Background(), filter, findOptions)
	for res.Next(context.TODO()) {
		err := res.Decode(&returnRecord)
		if err != nil {
			log.Print(err)
		}
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
