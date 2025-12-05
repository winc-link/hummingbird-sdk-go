package executors

import (
	"context"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"github.com/zeromicro/go-zero/core/executors"
	"time"
)

type TimeDataBatcher struct {
	executor *executors.ChunkExecutor
}

func NewTimeSeriesDataBatcher(dataDb datadb.DataBase) *TimeDataBatcher {
	return &TimeDataBatcher{
		executor: executors.NewChunkExecutor(
			func(tasks []any) {
				// 批量写入日志
				datas := make([]model.BatchInsertPropertyData, 0, len(tasks))

				for _, task := range tasks {
					datas = append(datas, task.(model.BatchInsertPropertyData))
				}
				// 一次性写入数据库
				dataDb.InsertBatch(context.Background(), datas)
			},
			executors.WithChunkBytes(1024*1024),         // 1MB 触发
			executors.WithFlushInterval(10*time.Second), // 或 1 秒触发
		),
	}
}

func (l *TimeDataBatcher) AddData(msg any) {
	l.executor.Add(msg, 1)
}
