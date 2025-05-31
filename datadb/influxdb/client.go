package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
)

type Client struct {
	org, bucket string
	client      influxdb2.Client
}

type DbClient struct {
	Org    string
	Bucket string
	Url    string
	Token  string
}

func (c *Client) Insert(ctx context.Context, table string, fields map[string]interface{}, t int64) (err error) {
	//get non-blocking write client
	//timestamp := time.Now()
	var ts time.Time
	ts = time.UnixMilli(t).UTC()
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	p := influxdb2.NewPoint(table,
		map[string]string{},
		fields,
		ts)
	// write point asynchronously
	writeAPI.WritePoint(p)
	// Flush writes
	writeAPI.Flush()
	return nil
}

func (c *Client) InsertPropertyData(ctx context.Context, tag map[string]string, fields map[string]interface{}, t int64) (err error) {
	var ts time.Time
	ts = time.UnixMilli(t).UTC()
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	p := influxdb2.NewPoint("property_data",
		tag,
		fields,
		ts)
	writeAPI.WritePoint(p)
	writeAPI.Flush()
	return nil
}

func (c *Client) InsertEventData(ctx context.Context, tag map[string]string, fields map[string]interface{}, t int64) (err error) {
	var ts time.Time
	ts = time.UnixMilli(t).UTC()
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	p := influxdb2.NewPoint("event_data",
		tag,
		fields,
		ts)
	writeAPI.WritePoint(p)
	writeAPI.Flush()
	return nil
}

func (c *Client) Close() {
	c.client.Close()
}
func InitClientInfluxDB(config DbClient) (datadb.DataBase, error) {
	client := influxdb2.NewClient(config.Url, config.Token)
	ok, err := client.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("influxdb2 ping failed")
	}
	return &Client{
		client: client,
		org:    config.Org,
		bucket: config.Bucket,
	}, nil
}
