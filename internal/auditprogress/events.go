package auditprogress

import (
	"encoding/json"
	"fmt"
	"ops-web/internal/logger"
	"sync"
	"time"
)

// 事件类型
const (
	EventTypeTaskCreated   = "task_created"   // 新任务导入
	EventTypeTaskUpdated   = "task_updated"   // 任务更新（审核状态、审核意见等）
	EventTypeTaskDeleted   = "task_deleted"   // 任务删除
	EventTypeTaskSampled   = "task_sampled"   // 任务抽检
	EventTypeTaskRefreshed = "task_refreshed" // 需要刷新整个列表
)

// Event 表示一个更新事件
type Event struct {
	Type      string                 `json:"type"`      // 事件类型
	TaskID    int                    `json:"task_id"`   // 任务ID（如果适用）
	Data      map[string]interface{} `json:"data"`      // 事件数据
	Timestamp time.Time              `json:"timestamp"` // 事件时间戳
}

// Client 表示一个SSE客户端连接
type Client struct {
	ID       string          // 客户端唯一ID
	Send     chan Event      // 发送事件的通道
	UserID   int             // 用户ID（用于权限控制）
	Page     int             // 当前页码（用于过滤）
	Filters  map[string]string // 当前筛选条件
}

// EventHub 管理所有SSE连接和事件广播
type EventHub struct {
	clients    map[string]*Client // 所有客户端连接
	broadcast   chan Event         // 广播通道
	register    chan *Client       // 注册新客户端
	unregister  chan *Client       // 注销客户端
	mu          sync.RWMutex       // 读写锁
}

// NewEventHub 创建新的事件中心
func NewEventHub() *EventHub {
	return &EventHub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Event, 256), // 缓冲256个事件
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动事件中心（在goroutine中运行）
func (h *EventHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			clientCount := len(h.clients)
			h.mu.Unlock()
			logger.Infof("[EventHub] 客户端连接: %s (用户ID: %d, 总数: %d)", client.ID, client.UserID, clientCount)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			logger.Infof("[EventHub] 客户端断开: %s (剩余: %d)", client.ID, len(h.clients))

		case event := <-h.broadcast:
			// 广播事件给所有客户端
			h.mu.RLock()
			clientCount := len(h.clients)
			clients := make([]*Client, 0, clientCount)
			for _, client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()

			// 发送事件给所有客户端
			logger.Infof("[EventHub] 开始广播事件给 %d 个客户端: type=%s, taskID=%d", clientCount, event.Type, event.TaskID)
			if clientCount == 0 {
				logger.Infof("[EventHub] 警告: 没有活跃的客户端连接，事件将被丢弃")
			}
			sentCount := 0
			for _, client := range clients {
				select {
				case client.Send <- event:
					sentCount++
				default:
					// 如果客户端通道已满，跳过（避免阻塞）
					logger.Infof("[EventHub] 客户端 %s 通道已满，跳过事件", client.ID)
				}
			}
			logger.Infof("[EventHub] 事件已发送给 %d/%d 个客户端", sentCount, clientCount)
		}
	}
}

// Broadcast 广播事件给所有客户端
func (h *EventHub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
		logger.Infof("[EventHub] 事件已加入广播队列: type=%s, taskID=%d", event.Type, event.TaskID)
	default:
		// 如果广播通道已满，记录警告
		logger.Infof("[EventHub] 警告: 广播通道已满，事件可能丢失: type=%s, taskID=%d", event.Type, event.TaskID)
	}
}

// BroadcastTaskCreated 广播新任务创建事件
func (h *EventHub) BroadcastTaskCreated(taskID int, taskData map[string]interface{}) {
	event := Event{
		Type:      EventTypeTaskCreated,
		TaskID:    taskID,
		Data:      taskData,
		Timestamp: time.Now(),
	}
	h.Broadcast(event)
}

// BroadcastTaskUpdated 广播任务更新事件
func (h *EventHub) BroadcastTaskUpdated(taskID int, updateData map[string]interface{}) {
	event := Event{
		Type:      EventTypeTaskUpdated,
		TaskID:    taskID,
		Data:      updateData,
		Timestamp: time.Now(),
	}
	logger.Infof("[EventHub] 广播任务更新事件: taskID=%d, data=%v", taskID, updateData)
	h.Broadcast(event)
}

// BroadcastTaskDeleted 广播任务删除事件
func (h *EventHub) BroadcastTaskDeleted(taskID int) {
	event := Event{
		Type:      EventTypeTaskDeleted,
		TaskID:    taskID,
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
	h.Broadcast(event)
}

// BroadcastTaskSampled 广播任务抽检事件
func (h *EventHub) BroadcastTaskSampled(taskID int, sampleData map[string]interface{}) {
	event := Event{
		Type:      EventTypeTaskSampled,
		TaskID:    taskID,
		Data:      sampleData,
		Timestamp: time.Now(),
	}
	h.Broadcast(event)
}

// BroadcastRefresh 广播刷新事件（用于需要刷新整个列表的情况）
func (h *EventHub) BroadcastRefresh() {
	event := Event{
		Type:      EventTypeTaskRefreshed,
		TaskID:    0,
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
	h.Broadcast(event)
}

// 全局事件中心实例
var globalEventHub *EventHub
var eventHubOnce sync.Once

// GetEventHub 获取全局事件中心实例（单例模式）
func GetEventHub() *EventHub {
	eventHubOnce.Do(func() {
		globalEventHub = NewEventHub()
		go globalEventHub.Run()
	})
	return globalEventHub
}

// RegisterClient 注册客户端（公共方法，供其他包使用）
func (h *EventHub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient 注销客户端（公共方法，供其他包使用）
func (h *EventHub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// FormatSSE 将事件格式化为SSE格式
func FormatSSE(event Event) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data: %s\n\n", string(data)), nil
}

