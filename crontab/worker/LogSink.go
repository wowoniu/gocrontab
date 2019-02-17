package worker

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"gocrontab/crontab/common"
)

type LogSink struct {
	Client     *mongo.Client
	DataBase   *mongo.Database
	Collection *mongo.Collection
	LogChan    chan *common.JobLog
}

var G_log *LogSink

func LoadLog() (err error) {
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

	G_log = &LogSink{
		Client:     client,
		DataBase:   database,
		Collection: collection,
		LogChan:    make(chan *common.JobLog, 1000),
	}
	//启动协程 监听日志channel
	go G_log.WriteLoop()
	return
}

func (this *LogSink) PushLog(logRecord *common.JobLog) {
	go func() {
		var (
			err error
			//bsonData []byte
		)
		//if bsonData,err=bson.Marshal(logRecord);err!=nil{
		//	return
		//}
		if _, err = this.Collection.InsertOne(context.TODO(), logRecord); err != nil {
			fmt.Println("日志写入失败:", err)
			return
		}
		//插入成功
		return
	}()
	return
}

func (this *LogSink) WriteLoop() {
	var (
		jobLog   *common.JobLog
		bsonData []byte
		err      error
	)
	for jobLog = range G_log.LogChan {
		if bsonData, err = bson.Marshal(jobLog); err != nil {
			continue
		}
		if _, err = G_log.Collection.InsertOne(context.TODO(), bsonData); err != nil {
			//fmt.Println("日志写入失败:",err)
		}
	}
}
