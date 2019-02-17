package master

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"gocrontab/crontab/common"
)

type JobLog struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

var G_joblog *JobLog

func LoadJobLog() (err error) {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
	)
	if client, err = mongo.Connect(context.TODO(), G_config.MongoHost); err != nil {
		return
	}
	database = client.Database("cron")
	collection = database.Collection("log")

	G_joblog = &JobLog{
		Client:     client,
		Database:   database,
		Collection: collection,
	}
	return
}

func (this *JobLog) GetLogList(jobName string, page int64, pageSize int64) (batchLogs []*common.JobLog, count int64, err error) {
	var (
		findCursor  *mongo.Cursor
		jobLog      *common.JobLog
		findOptions *options.FindOptions
		skip        int64
	)
	skip = (page - 1) * pageSize
	batchLogs = make([]*common.JobLog, 0)

	findOptions = &options.FindOptions{}
	findOptions.SetSkip(skip)
	findOptions.SetLimit(pageSize)
	findOptions.SetSort(bson.M{"_id": -1})
	if findCursor, err = this.Collection.Find(context.TODO(), bson.M{"job_name": jobName}, findOptions); err != nil {
		return
	}
	//统计
	if count, err = this.Collection.CountDocuments(context.TODO(), bson.M{"job_name": jobName}); err != nil {
		err = nil
		count = 0
	}
	defer findCursor.Close(context.TODO())
	for findCursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}
		if err = findCursor.Decode(jobLog); err != nil {
			continue
		}
		batchLogs = append(batchLogs, jobLog)
	}

	return
}
