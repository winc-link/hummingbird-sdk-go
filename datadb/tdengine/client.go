package tdengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/taosdata/driver-go/v3/taosWS"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"sort"
	"strings"
)

type Client struct {
	client *sql.DB
}

func (c *Client) InsertBatchDeviceEvent(ctx context.Context, p model.BatchInsertEventData) error {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	first := true

	if p.DeviceID == "" {
		return fmt.Errorf("deviceId is empty")
	}
	// 表名，例如 device_123
	table := fmt.Sprintf(constants.TdengineDataDB_PREFIX+"%s", p.DeviceID)

	if !first {
		sb.WriteString(" ")
	}
	first = false

	sb.WriteString(table)

	// 写入列名：ts + Data 的 key
	sb.WriteString(" (ts,event_type")

	keys := mapKeysInOrder(p.Data)

	for _, key := range keys {
		sb.WriteString(",")
		sb.WriteString(key)
	}
	sb.WriteString(") VALUES (")

	// 写入时间戳
	sb.WriteString(fmt.Sprintf("%d", p.T))
	sb.WriteString(",")
	sb.WriteString(p.EventType)

	// 按 Data map 的 key 写入值
	for _, key := range keys {
		v := p.Data[key]
		sb.WriteString(",")
		sb.WriteString(formatValue(v))
	}

	sb.WriteString(")")

	_, err := c.client.Exec(sb.String())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) InsertBatchDeviceEvents(ctx context.Context, points []model.BatchInsertEventData) error {
	if len(points) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	first := true

	for _, p := range points {
		if p.DeviceID == "" {
			return fmt.Errorf("deviceId is empty")
		}
		// 表名，例如 device_123
		table := fmt.Sprintf(constants.TdengineDataDB_PREFIX+"%s", p.DeviceID)

		if !first {
			sb.WriteString(" ")
		}
		first = false

		sb.WriteString(table)

		// 写入列名：ts + Data 的 key
		sb.WriteString(" (ts,event_type")

		keys := mapKeysInOrder(p.Data)

		for _, key := range keys {
			sb.WriteString(",")
			sb.WriteString(key)
		}
		sb.WriteString(") VALUES (")

		// 写入时间戳
		sb.WriteString(fmt.Sprintf("%d", p.T))
		sb.WriteString(",")
		sb.WriteString(p.EventType)

		// 按 Data map 的 key 写入值
		for _, key := range keys {
			v := p.Data[key]
			sb.WriteString(",")
			sb.WriteString(formatValue(v))
		}

		sb.WriteString(")")
	}

	_, err := c.client.Exec(sb.String())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) InsertBatchDeviceLogs(ctx context.Context, logs []model.BatchInsertDeviceLogData) error {
	if len(logs) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	first := true

	for _, p := range logs {
		if p.DeviceID == "" {
			return fmt.Errorf("deviceId is empty")
		}
		// 表名，例如 device_123
		table := fmt.Sprintf(constants.TdengineLogDB_PREFIX+"%s", p.DeviceID)

		if !first {
			sb.WriteString(" ")
		}
		first = false

		sb.WriteString(table)

		// 写入列名：ts + Data 的 key
		sb.WriteString(" (ts")

		dataMap := model.CovertLogDataToMap(p.Data)
		keys := mapKeysInOrder(dataMap)

		for _, key := range keys {
			sb.WriteString(",")
			sb.WriteString(key)
		}
		sb.WriteString(") VALUES (")

		// 写入时间戳
		sb.WriteString(fmt.Sprintf("%d", p.T))
		// 写值
		for _, key := range keys {
			v := dataMap[key]
			sb.WriteString(",")
			sb.WriteString(formatValue(v))
		}
		sb.WriteString(")")
	}
	_, err := c.client.Exec(sb.String())
	if err != nil {
		return err
	}

	return nil
}

type DbClient struct {
	Dsn string
}

func (c *Client) InsertDeviceProperties(ctx context.Context, p model.BatchInsertPropertyData) error {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	first := true

	if p.DeviceID == "" {
		return fmt.Errorf("deviceId is empty")
	}
	// 表名，例如 device_123
	table := fmt.Sprintf(constants.TdengineDataDB_PREFIX+"%s", p.DeviceID)

	if !first {
		sb.WriteString(" ")
	}
	first = false

	sb.WriteString(table)

	// 写入列名：ts + Data 的 key
	sb.WriteString(" (ts")

	keys := mapKeysInOrder(p.Data)

	for _, key := range keys {
		sb.WriteString(",")
		sb.WriteString(key)
	}
	sb.WriteString(") VALUES (")

	// 写入时间戳
	sb.WriteString(fmt.Sprintf("%d", p.T))

	// 按 Data map 的 key 写入值
	for _, key := range keys {
		v := p.Data[key]
		sb.WriteString(",")
		sb.WriteString(formatValue(v))
	}

	sb.WriteString(")")

	_, err := c.client.Exec(sb.String())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) InsertBatchDeviceProperties(ctx context.Context, points []model.BatchInsertPropertyData) error {
	if len(points) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	first := true

	for _, p := range points {
		if p.DeviceID == "" {
			return fmt.Errorf("deviceId is empty")
		}
		// 表名，例如 device_123
		table := fmt.Sprintf(constants.TdengineDataDB_PREFIX+"%s", p.DeviceID)

		if !first {
			sb.WriteString(" ")
		}
		first = false

		sb.WriteString(table)

		// 写入列名：ts + Data 的 key
		sb.WriteString(" (ts")

		keys := mapKeysInOrder(p.Data)

		for _, key := range keys {
			sb.WriteString(",")
			sb.WriteString(key)
		}
		sb.WriteString(") VALUES (")

		// 写入时间戳
		sb.WriteString(fmt.Sprintf("%d", p.T))

		// 按 Data map 的 key 写入值
		for _, key := range keys {
			v := p.Data[key]
			sb.WriteString(",")
			sb.WriteString(formatValue(v))
		}

		sb.WriteString(")")
	}
	fmt.Println(sb.String())

	_, err := c.client.Exec(sb.String())
	if err != nil {
		return err
	}

	return nil
}

// 保证 map key 遍历顺序一致（可选）
func mapKeysInOrder(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // 按字母顺序排序
	return keys
}

// 格式化值
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case float64, float32, int, int32, int64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		b, _ := json.Marshal(val)
		return "'" + string(b) + "'"
	}
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
