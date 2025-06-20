package tdengine

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"strings"
	"time"
)

type Client struct {
	client *sql.DB
}

type DbClient struct {
	Dsn string
}

func (c *Client) Insert(ctx context.Context, table string, data map[string]interface{}, t int64) error {
	ts := time.UnixMilli(t) // 直接使用毫秒时间戳创建 time.Time

	var columns []string
	var placeholders []string
	var values []interface{}

	columns = append(columns, "ts")
	placeholders = append(placeholders, "?")
	values = append(values, ts) // 这里用 time.Time 类型，不要转成字符串

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		table,
		strings.Join(wrapWithBackticks(columns), ","),
		strings.Join(placeholders, ","),
	)
	_, err := c.client.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert data into %s: %v", table, err)
	}
	return nil
}

func wrapWithBackticks(fields []string) []string {
	for i, f := range fields {
		fields[i] = fmt.Sprintf("`%s`", f)
	}
	return fields
}

func (c *Client) Close() {
	c.client.Close()
}

func InitTDengineClient(config DbClient) (datadb.DataBase, error) {
	taos, err := sql.Open("taosWS", config.Dsn)
	if err != nil {
		return nil, err
	}
	// SetMaxOpenConns sets the maximum number of open connections to the database. 0 means unlimited.
	taos.SetMaxOpenConns(0)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	//taos.SetMaxIdleConns(20)

	err = taos.Ping()
	if err != nil {
		return nil, err
	}
	return &Client{client: taos}, nil
}
