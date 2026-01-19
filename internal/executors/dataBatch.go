package executors

import (
	"context"
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"github.com/winc-link/hummingbird-sdk-go/internal/logger"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"github.com/zeromicro/go-zero/core/executors"
	"time"
)

type PropertiesTimeDataBatcher struct {
	executor *executors.ChunkExecutor
}

func NewTimeDevicePropertiesDataBatcher(dataDb datadb.DataBase, log logger.Logger, customParam string) *PropertiesTimeDataBatcher {
	var (
		dataBatchSize     = 100 //100个
		dataBatchInterval = 10  //10s
	)
	customParamMap := make(map[string]interface{})
	if customParam != "" {
		err := json.Unmarshal([]byte(customParam), &customParamMap)
		if err != nil {
			log.Errorf("custom param parse error: %s", err.Error())
		}
	}
	if batchSize, ok := customParamMap[constants.DataBatchSize]; ok {
		if cast.ToInt(batchSize) > 0 {
			dataBatchSize = cast.ToInt(batchSize)
		}
	}
	if batchInterval, ok := customParamMap[constants.DataBatchInterval]; ok {
		if cast.ToInt(batchInterval) > 0 {
			dataBatchInterval = cast.ToInt(batchInterval)
		}
	}

	log.Infof("batch report dataBatchSize :%d dataBatchInterval:%d", dataBatchSize, dataBatchInterval)

	return &PropertiesTimeDataBatcher{
		executor: executors.NewChunkExecutor(
			func(tasks []any) {
				// 批量写入
				data := make([]model.BatchInsertPropertyData, 0, len(tasks))
				for _, task := range tasks {
					data = append(data, task.(model.BatchInsertPropertyData))
				}
				log.Infof("data:%v", data)
				// 一次性写入数据库
				err := dataDb.InsertBatchDeviceProperties(context.Background(), data)
				if err != nil {
					log.Error("Insert batch failed:", err)
				}
			},
			executors.WithChunkBytes(dataBatchSize),
			executors.WithFlushInterval(time.Duration(dataBatchInterval)*time.Second),
		),
	}
}
func (l *PropertiesTimeDataBatcher) AddData(msg any) {
	_ = l.executor.Add(msg, 1)
}
