package datadb

import (
	"context"
	"github.com/winc-link/hummingbird-sdk-go/model"
)

type DataBase interface {
	Insert(ctx context.Context, table string, fields map[string]interface{}, t int64) (err error)
	InsertBatch(ctx context.Context, points []model.BatchInsertPropertyData) error
	Close()
}
