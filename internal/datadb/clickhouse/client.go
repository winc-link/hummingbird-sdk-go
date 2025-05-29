package clickhouse

import (
	"context"
	//"crypto/tls"
	"fmt"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	"github.com/winc-link/hummingbird-sdk-go/internal/datadb"

	"github.com/ClickHouse/clickhouse-go/v2"
	//"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouse struct {
	conn clickhouse.Conn
}

func (c *ClickHouse) Insert(ctx context.Context, table string, fields map[string]interface{}, t int64) (err error) {
	//TODO implement me
	panic("implement me")
}

func (c *ClickHouse) Close() {
	return
}

func InitClientHouseClient(config *drivercommon.ClickHouse) (datadb.DataBase, error) {

	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: config.Addr,
			Auth: clickhouse.Auth{
				Database: config.Database,
				Username: config.Username,
				Password: config.Password,
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return &ClickHouse{
		conn: conn,
	}, nil
}
