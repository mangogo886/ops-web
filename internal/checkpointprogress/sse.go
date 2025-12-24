package checkpointprogress

import (
	"fmt"
	"net/http"
	"ops-web/internal/auth"
	"ops-web/internal/auditprogress"
	"ops-web/internal/logger"
	"strconv"
	"time"
)

// SSEHandler: Server-Sent Events 端点，用于实时推送更新
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	// 检查用户是否已登录
	currentUser := auth.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲

	// 创建客户端（使用auditprogress的Client类型）
	clientID := fmt.Sprintf("%d_%d", currentUser.ID, time.Now().UnixNano())
	client := &auditprogress.Client{
		ID:      clientID,
		Send:    make(chan auditprogress.Event, 256),
		UserID:  currentUser.ID,
		Page:    1, // 可以从查询参数获取
		Filters: make(map[string]string),
	}

	// 从查询参数获取筛选条件
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			client.Page = page
		}
	}
	if searchName := r.URL.Query().Get("file_name"); searchName != "" {
		client.Filters["file_name"] = searchName
	}
	if auditStatus := r.URL.Query().Get("audit_status"); auditStatus != "" {
		client.Filters["audit_status"] = auditStatus
	}
	if archiveType := r.URL.Query().Get("archive_type"); archiveType != "" {
		client.Filters["archive_type"] = archiveType
	}
	if sampleStatus := r.URL.Query().Get("sample_status"); sampleStatus != "" {
		client.Filters["sample_status"] = sampleStatus
	}

	// 注册客户端
	hub := auditprogress.GetEventHub()
	hub.RegisterClient(client)
	logger.Infof("[CheckpointSSEHandler] 客户端已注册: %s (用户ID: %d)", clientID, currentUser.ID)
	
	// 等待一小段时间确保注册完成
	time.Sleep(100 * time.Millisecond)

	// 确保在函数退出时注销客户端
	defer func() {
		logger.Infof("[CheckpointSSEHandler] 客户端断开连接: %s", clientID)
		hub.UnregisterClient(client)
	}()

	// 发送初始连接成功消息
	initialEvent := auditprogress.Event{
		Type:      "connected",
		TaskID:    0,
		Data:      map[string]interface{}{"message": "连接成功"},
		Timestamp: time.Now(),
	}
	if sseData, err := auditprogress.FormatSSE(initialEvent); err == nil {
		if _, err := fmt.Fprintf(w, sseData); err != nil {
			logger.Errorf("[CheckpointSSEHandler] 发送初始消息失败: %v", err)
			return
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		} else {
			logger.Infof("[CheckpointSSEHandler] 警告: ResponseWriter不支持Flush")
		}
		logger.Infof("[CheckpointSSEHandler] 初始连接消息已发送")
	} else {
		logger.Errorf("[CheckpointSSEHandler] 格式化初始消息失败: %v", err)
		return
	}

	// 监听客户端断开连接
	ctx := r.Context()
	notify := ctx.Done()

	// 发送心跳（每30秒）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 主循环：监听事件和心跳
	for {
		select {
		case <-notify:
			// 客户端断开连接
			return

		case event, ok := <-client.Send:
			if !ok {
				// 通道已关闭
				return
			}

			// 格式化并发送事件
			sseData, err := auditprogress.FormatSSE(event)
			if err != nil {
				logger.Errorf("SSE格式化失败: %v", err)
				continue
			}

			if _, err := fmt.Fprintf(w, sseData); err != nil {
				logger.Errorf("SSE发送失败: %v", err)
				return
			}

			// 立即刷新响应
			w.(http.Flusher).Flush()

		case <-ticker.C:
			// 发送心跳保持连接
			heartbeat := auditprogress.Event{
				Type:      "heartbeat",
				TaskID:    0,
				Data:      map[string]interface{}{},
				Timestamp: time.Now(),
			}
			if sseData, err := auditprogress.FormatSSE(heartbeat); err == nil {
				fmt.Fprintf(w, sseData)
				w.(http.Flusher).Flush()
			}
		}
	}
}

