package influxdb

import (
	"context"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
)

type Client struct {
	org, bucket, logBucket string
	client                 influxdb2.Client
}

type DbClient struct {
	Org       string
	Bucket    string //存储设备上报数据
	LogBucket string //存储设备日志信息
	Url       string
	Token     string
}

func (c *Client) InsertDeviceProperties(ctx context.Context, p model.BatchInsertPropertyData) error {
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	ts := time.UnixMilli(p.T).UTC()
	point := influxdb2.NewPoint(
		"device_properties",
		map[string]string{
			"device_id": p.DeviceID,
		},
		p.Data,
		ts,
	)
	writeAPI.WritePoint(point)
	// 刷盘，确保写入完成
	writeAPI.Flush()
	return nil
}

func (c *Client) InsertBatchDeviceProperties(ctx context.Context, points []model.BatchInsertPropertyData) error {
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	for _, p := range points {
		fmt.Println("DeviceID:", p.DeviceID)
		ts := time.UnixMilli(p.T).UTC()
		point := influxdb2.NewPoint(
			"device_properties",
			map[string]string{
				"device_id": p.DeviceID,
			},
			p.Data,
			ts,
		)
		writeAPI.WritePoint(point)
		// 刷盘，确保写入完成
		writeAPI.Flush()
	}
	return nil
}

//func buildPoints() []*influxdb2.Point {
//	now := time.Now()
//
//	points := make([]*influxdb2.Point, 0, 100)
//
//	for i := 0; i < 100; i++ {
//		p := influxdb2.NewPoint(
//			"device_property",
//			map[string]string{
//				"deviceId": "dev-001",
//				"metric":   "temperature",
//			},
//			map[string]interface{}{
//				"value": 23.5 + float64(i)*0.1,
//			},
//			now.Add(time.Duration(i)*time.Second),
//		)
//		points = append(points, p)
//	}
//	return points
//}

func (c *Client) InsertBatchDeviceEvent(ctx context.Context, p model.BatchInsertEventData) error {
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	ts := time.UnixMilli(p.T).UTC()
	point := influxdb2.NewPoint(
		"device_events",
		map[string]string{
			"device_id":  p.DeviceID,
			"event_type": p.EventType,
		},
		p.Data,
		ts,
	)
	writeAPI.WritePoint(point)
	// 刷盘，确保写入完成
	writeAPI.Flush()
	return nil
}

func (c *Client) InsertBatchDeviceEvents(ctx context.Context, points []model.BatchInsertEventData) error {
	writeAPI := c.client.WriteAPI(c.org, c.bucket)
	for _, p := range points {
		ts := time.UnixMilli(p.T).UTC()
		point := influxdb2.NewPoint(
			"device_events",
			map[string]string{
				"device_id":  p.DeviceID,
				"event_type": p.EventType,
			},
			p.Data,
			ts,
		)
		writeAPI.WritePoint(point)
	}
	// 刷盘，确保写入完成
	writeAPI.Flush()
	return nil
}

func (c *Client) InsertBatchDeviceLogs(ctx context.Context, points []model.BatchInsertDeviceLogData) error {
	writeAPI := c.client.WriteAPI(c.org, c.logBucket)
	for _, p := range points {
		ts := time.UnixMilli(p.T).UTC()
		point := influxdb2.NewPoint(
			"device-log",
			map[string]string{
				"device_id": p.DeviceID,
				"log_type":  string(p.Data.LogType),
			},
			model.CovertLogDataToMap(p.Data),
			ts,
		)
		writeAPI.WritePoint(point)
	}
	// 刷盘，确保写入完成
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
		client:    client,
		org:       config.Org,
		bucket:    config.Bucket,
		logBucket: config.LogBucket,
	}, nil
}
