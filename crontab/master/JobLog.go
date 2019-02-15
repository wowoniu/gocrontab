package master

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
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
	if client, err = mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017"); err != nil {
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

func (this *JobLog) GetLogList(jobName string) (batchLogs []*common.JobLog, err error) {
	var (
		findCursor *mongo.Cursor
		jobLog     *common.JobLog
	)
	batchLogs = make([]*common.JobLog, 0)
	if findCursor, err = this.Collection.Find(context.TODO(), bson.M{"JobName": jobName}); err != nil {
		return
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
