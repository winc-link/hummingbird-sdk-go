package tdengine

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	"github.com/winc-link/hummingbird-sdk-go/internal/datadb"
	"strings"
	"time"
)

type Client struct {
	client *sql.DB
}

func (c *Client) Insert(ctx context.Context, table string, data map[string]interface{}, t int64) (err error) {
	ts := time.Unix(0, t*int64(time.Millisecond))
	formattedTime := ts.UTC().Format("2006-01-02 15:04:05.000")
	var (
		field = []string{"ts"}
		value = []string{"'" + formattedTime + "'"}
	)

	for k, v := range data {
		field = append(field, strings.ToLower(k))
		switch val := v.(type) {
		case string:
			value = append(value, "'"+val+"'")
		case int, int64, float64, float32:
			value = append(value, fmt.Sprintf("%v", val))
		default:
			value = append(value, "null")
		}
	}
	s := "INSERT INTO ? (?) VALUES (?)"
	_, err = c.client.ExecContext(ctx, s, table, strings.Join(field, ","), strings.Join(value, ","))
	return
}

func (c *Client) Close() {
	c.client.Close()
}

func InitTDengineClient(config *drivercommon.TDengineSource) (datadb.DataBase, error) {
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
