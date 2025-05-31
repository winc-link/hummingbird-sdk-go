package datadb

import "context"

type DataBase interface {
	Insert(ctx context.Context, table string, fields map[string]interface{}, t int64) (err error)
	Close()
}
