package monitor

import (
	"sync"
	"sync/atomic"
	"time"
)

type QPSRecord struct {
	T       string `json:"t"`
	UpQPS   uint64 `json:"UpQps"`
	DownQPS uint64 `json:"DownQps"`
}

var (
	UpReqCount   uint64
	DownReqCount uint64
	history      []QPSRecord // 缓存的 QPS 记录
	historyMux   sync.Mutex  // 保护 history 线程安全
	historyCap   = 120       // 最多保留 120 条记录

)

func UpQosRequest() {
	atomic.AddUint64(&UpReqCount, 1)
}

func DownQosRequest() {
	atomic.AddUint64(&DownReqCount, 1)
}

// 每秒收集一次 QPS 并加入缓存
func StartQPSCollector() {
	initHistory(historyCap)
	ticker := time.NewTicker(time.Second)
	go func() {
		for t := range ticker.C {
			upQps := atomic.SwapUint64(&UpReqCount, 0)
			downQps := atomic.SwapUint64(&DownReqCount, 0)

			record := QPSRecord{
				T:       t.Format("15:04:05"),
				UpQPS:   upQps,
				DownQPS: downQps,
			}

			historyMux.Lock()
			history = append(history, record)
			if len(history) > historyCap {
				history = history[1:] // 删除最老的一条
			}
			historyMux.Unlock()

		}
	}()
}

func initHistory(cap int) {
	historyMux.Lock()
	defer historyMux.Unlock()
	now := time.Now().Add(-time.Duration(cap) * time.Second)
	for i := 0; i < cap; i++ {
		history = append(history, QPSRecord{
			T:       now.Add(time.Duration(i) * time.Second).Format("15:04:05"),
			UpQPS:   0,
			DownQPS: 0,
		})
	}
}

func GetQosHistory() []QPSRecord {
	historyMux.Lock()
	defer historyMux.Unlock()
	return history
}
