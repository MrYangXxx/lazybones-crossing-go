package service

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	"hidevops.io/hiboot-data/starter/mongo"
	"hidevops.io/hiboot/pkg/app"
	"lazybones-crossing-go/entity"
	"log"
)

type RecordService interface {
	AddRecord(record *entity.Record) error
	FindByFilter(*entity.Record, int64, int64) (records *[]entity.Record, pagination *entity.Pagination, err error)
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

func (r *recordServiceImpl) AddRecord(record *entity.Record) error {
	if record == nil {
		return errors.New("user is not allowed nil")
	}
	db := r.client.Database("lazybones").Collection("records")
	_, err := db.InsertOne(context.Background(), record)
	return err
}

func (r *recordServiceImpl) FindByFilter(record *entity.Record, page int64, pageSize int64) (records *[]entity.Record, pagination *entity.Pagination, err error) {
	records = &[]entity.Record{}
	returnRecord := entity.Record{}

	db := r.client.Database("lazybones").Collection("records")
	skip := (page - 1) * pageSize
	findOptions := options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
	}

	res, err := db.Find(context.Background(), record, &findOptions)
	for res.Next(context.TODO()) {
		err := res.Decode(&returnRecord)
		if err != nil {
			log.Print(err)
		}
		*records = append(*records, returnRecord)
	}
	count, err := db.CountDocuments(context.Background(), record)
	pagination = &entity.Pagination{
		Page:     int(page),
		PageSize: int(pageSize),
		Count:    int(count),
	}

	return
}
