package clickhouse

import (
	"context"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/internal/datadb"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type DbClient struct {
	Addr     []string
	Database string
	Username string
	Password string
}

type ClickHouse struct {
	conn clickhouse.Conn
}

func (c *ClickHouse) Insert(ctx context.Context, table string, fields map[string]interface{}, t int64) (err error) {
	if len(fields) == 0 {
		return fmt.Errorf("fields cannot be empty")
	}

	// 自动补充 timestamp 字段
	fields["timestamp"] = time.UnixMilli(t)

	// 提取字段和占位符
	columns := make([]string, 0, len(fields))
	placeholders := make([]string, 0, len(fields))
	values := make([]interface{}, 0, len(fields))

	for k, v := range fields {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	// 构造 SQL
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// 执行插入
	return c.conn.Exec(ctx, query, values...)
}

func (c *ClickHouse) Close() {
	return
}

func InitClientHouseClient(config DbClient) (datadb.DataBase, error) {

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
