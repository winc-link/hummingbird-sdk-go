package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"strconv"
	"time"
)

var streamKey = "iot-stream"

type Client struct {
	ctx    context.Context
	client *redis.Client
}

func (c *Client) Ping() error {
	return c.client.Ping(context.Background()).Err()
}

// PushMsgToStream 向redis Stream推送数据，等待下游消费
func (c *Client) PushMsgToStream(data []byte) error {
	_, err := c.client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: streamKey,
		ID:     "*", // 自动生成消息ID
		Values: data,
	}).Result()

	if err != nil {
		return err
	}
	return nil
}

// 保存设备状态到 Redis（Hash结构）
func (c *Client) SaveDeviceState(deviceID string, data map[string]interface{}) error {
	key := fmt.Sprintf("device:%s", deviceID)
	return c.client.HSet(context.Background(), key, data).Err()
}

// 读取设备单个属性
func (c *Client) GetDeviceField(deviceID, field string) (string, error) {
	key := fmt.Sprintf("device:%s", deviceID)
	return c.client.HGet(context.Background(), key, field).Result()
}

// 读取设备所有属性
func (c *Client) GetDeviceState(deviceID string) (map[string]string, error) {
	key := fmt.Sprintf("device:%s", deviceID)
	return c.client.HGetAll(context.Background(), key).Result()
}

func (c *Client) GetDeviceFields(deviceID string, fields []string) (map[string]string, error) {
	key := fmt.Sprintf("device:%s", deviceID)

	// 如果字段为空，则获取全部属性
	if len(fields) == 0 {
		return c.client.HGetAll(context.Background(), key).Result()
	}
	// 否则使用 HMGET 获取部分字段
	values, err := c.client.HMGet(context.Background(), key, fields...).Result()
	if err != nil {
		return nil, err
	}
	// 组装为 map[string]string
	result := make(map[string]string, len(fields))
	for i, f := range fields {
		if values[i] != nil {
			result[f] = fmt.Sprintf("%v", values[i])
		}
	}

	return result, nil
}

//----------------------------

// String 字符串表示

// UpdateDeviceData 更新设备数据（自动记录各字段时间戳）
func (c *Client) UpdateDeviceData(deviceID string, reportTime int64, data map[string]interface{}) error {
	// 开启管道批量操作
	pipe := c.client.Pipeline()
	// 1. 更新主数据
	if len(data) > 0 {
		mainData := make(map[string]interface{})
		for k, v := range data {
			mainData[k] = v
		}
		mainData["last_update_ts"] = strconv.FormatInt(reportTime, 10)

		pipe.HSet(c.ctx, getMainDataKey(deviceID), mainData)
	}
	// 2. 更新各字段时间戳
	for field := range data {
		pipe.HSet(c.ctx, getTimestampsKey(deviceID), field, reportTime)
	}
	// 执行管道操作
	_, err := pipe.Exec(c.ctx)
	return err
}

// UpdateSingleField 更新单个字段
func (c *Client) UpdateSingleField(deviceID string, reportTime int64, field string, value interface{}) error {
	data := map[string]interface{}{field: value}
	return c.UpdateDeviceData(deviceID, reportTime, data)
}

// GetDeviceData 获取设备主数据
func (c *Client) GetDeviceData(deviceID string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, getMainDataKey(deviceID)).Result()
}

// GetDeviceDataWithTimestamps 获取设备数据及所有时间戳
func (c *Client) GetDeviceDataWithTimestamps(deviceID string) (*model.DeviceDataWithTimestamps, error) {
	// 使用管道批量获取
	pipe := c.client.Pipeline()
	mainDataCmd := pipe.HGetAll(c.ctx, getMainDataKey(deviceID))
	timestampsCmd := pipe.HGetAll(c.ctx, getTimestampsKey(deviceID))

	_, err := pipe.Exec(c.ctx)
	if err != nil {
		return nil, err
	}

	mainData, err := mainDataCmd.Result()
	if err != nil {
		return nil, err
	}

	timestamps, err := timestampsCmd.Result()
	if err != nil {
		return nil, err
	}

	// 转换时间戳为int64
	timestampInts := make(map[string]int64)
	for field, tsStr := range timestamps {
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			timestampInts[field] = ts
		}
	}

	// 解析最后更新时间
	var lastUpdate time.Time
	if lastUpdateStr, exists := mainData["last_update_ts"]; exists {
		if ts, err := strconv.ParseInt(lastUpdateStr, 10, 64); err == nil {
			lastUpdate = time.Unix(ts, 0)
		}
	}

	result := &model.DeviceDataWithTimestamps{
		DeviceID:   deviceID,
		Data:       mainData,
		Timestamps: timestampInts,
		LastUpdate: lastUpdate,
	}

	return result, nil
}

// GetFieldWithTimestamp 获取特定字段及其时间戳
func (c *Client) GetFieldWithTimestamp(deviceID, field string) (*model.FieldWithTimestamp, error) {
	pipe := c.client.Pipeline()
	valueCmd := pipe.HGet(c.ctx, getMainDataKey(deviceID), field)
	timestampCmd := pipe.HGet(c.ctx, getTimestampsKey(deviceID), field)

	_, err := pipe.Exec(c.ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	value, err := valueCmd.Result()
	if err != nil {
		return nil, err
	}

	timestampStr, err := timestampCmd.Result()
	if err != nil {
		return nil, err
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &model.FieldWithTimestamp{
		Field:     field,
		Value:     value,
		Timestamp: timestamp,
		HumanTime: time.Unix(timestamp, 0),
	}, nil
}

func (c *Client) GetFieldsWithTimestamps(deviceID string, fields []string) (map[string]*model.FieldWithTimestamp, error) {
	if len(fields) == 0 {
		return make(map[string]*model.FieldWithTimestamp), nil
	}

	result := make(map[string]*model.FieldWithTimestamp)
	mainKey := getMainDataKey(deviceID)
	timestampKey := getTimestampsKey(deviceID)

	for _, field := range fields {
		// 不使用管道，直接查询
		value, valueErr := c.client.HGet(c.ctx, mainKey, field).Result()
		timestampStr, timestampErr := c.client.HGet(c.ctx, timestampKey, field).Result()

		//fmt.Printf("直接查询 - field: %s, value: '%s', valueErr: %v\n", field, value, valueErr)
		//fmt.Printf("直接查询 - field: %s, timestamp: '%s', timestampErr: %v\n", field, timestampStr, timestampErr)

		if valueErr == redis.Nil || timestampErr == redis.Nil {
			//fmt.Printf("直接查询也返回 Nil - 字段 %s 可能真的不存在\n", field)
			continue
		}

		if valueErr != nil {
			continue
			//return nil, fmt.Errorf("读取字段 %s 的值失败: %w", field, valueErr)
		}
		if timestampErr != nil {
			continue
			//return nil, fmt.Errorf("读取字段 %s 的时间戳失败: %w", field, timestampErr)
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("解析字段 %s 的时间戳失败: %w", field, err)
		}

		result[field] = &model.FieldWithTimestamp{
			Field:     field,
			Value:     value,
			Timestamp: timestamp,
			HumanTime: time.Unix(timestamp, 0),
		}
	}
	return result, nil
}

// 辅助方法
func getMainDataKey(deviceID string) string {
	return fmt.Sprintf("device:%s", deviceID)
}

func getTimestampsKey(deviceID string) string {
	return fmt.Sprintf("device:%s:timestamps", deviceID)
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func NewClient(address, password string, db int) (c *Client, err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,  // Redis 地址
		Password: password, // 无密码则留空
		DB:       db,       // 使用默认数据库
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client: rdb,
		ctx:    context.Background(),
	}, nil
}
