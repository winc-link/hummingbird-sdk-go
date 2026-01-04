package datadb

import (
	"context"
	"github.com/winc-link/hummingbird-sdk-go/model"
)

type DataBase interface {
	InsertDeviceProperties(ctx context.Context, p model.BatchInsertPropertyData) error
	InsertBatchDeviceProperties(ctx context.Context, points []model.BatchInsertPropertyData) error
	InsertBatchDeviceEvent(ctx context.Context, p model.BatchInsertEventData) error
	InsertBatchDeviceEvents(ctx context.Context, points []model.BatchInsertEventData) error
	InsertBatchDeviceLogs(ctx context.Context, points []model.BatchInsertDeviceLogData) error
	Close()
}
