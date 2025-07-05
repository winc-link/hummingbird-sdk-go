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
	ts := time.UnixMilli(t)
	var columns []string
	var values []string
	// 添加时间字段
	columns = append(columns, "`ts`")
	values = append(values, fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05.000")))
	// 遍历字段
	for key, val := range data {
		columns = append(columns, fmt.Sprintf("`%s`", key))
		values = append(values, fmt.Sprintf("'%v'", escapeSQLString(val)))
	}
	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	)
	// 执行构造好的完整 SQL 字符串
	_, err := c.client.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to insert into %s: %v", table, err)
	}
	return nil
}

func escapeSQLString(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	// 转义单引号：' → ''
	return strings.ReplaceAll(s, "'", "''")
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
