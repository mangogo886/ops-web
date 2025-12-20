package auditprogress

import (
	"fmt"
	"ops-web/internal/logger"
	"time"
)

// StartVideoReminderScheduler 启动录像提醒定时任务
// 根据数据库配置执行定时任务
func StartVideoReminderScheduler() {
	go func() {
		for {
			// 获取定时配置
			config, err := GetScheduleConfig()
			if err != nil {
				logger.Errorf("获取定时配置失败: %v，使用默认配置（每天凌晨1点）", err)
				config = &ScheduleConfig{
					Frequency: "daily",
					Hour:      1,
					Enabled:   true,
				}
			}

			if !config.Enabled {
				logger.Errorf("定时任务已禁用，等待60秒后重新检查配置")
				time.Sleep(60 * time.Second)
				continue
			}

			// 计算下次执行时间
			nextRun := calculateNextRunTime(config)
			now := time.Now()
			waitDuration := nextRun.Sub(now)

			if waitDuration > 0 {
				logger.Errorf("定时任务已配置：%s，下次执行时间：%s，等待 %v", 
					getScheduleDescription(config), nextRun.Format("2006-01-02 15:04:05"), waitDuration)
				time.Sleep(waitDuration)
			}

			// 执行任务
			logger.Errorf("开始执行设备审核录像提醒定时任务（配置：%s）", getScheduleDescription(config))
			if err := ProcessVideoReminders(); err != nil {
				logger.Errorf("执行设备审核录像提醒定时任务失败: %v", err)
			}

			// 如果是每天执行，等待24小时；如果是每周执行，等待7天
			if config.Frequency == "daily" {
				time.Sleep(24 * time.Hour)
			} else {
				time.Sleep(7 * 24 * time.Hour)
			}
		}
	}()
	logger.Errorf("设备审核录像提醒定时任务已启动，将根据配置执行")
}

// calculateNextRunTime 计算下次执行时间
func calculateNextRunTime(config *ScheduleConfig) time.Time {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), config.Hour, 0, 0, 0, now.Location())

	if config.Frequency == "daily" {
		// 每天执行
		if today.After(now) || today.Equal(now) {
			// 今天还没到执行时间或正好是执行时间
			return today
		}
		// 今天已经过了执行时间，明天执行
		return today.Add(24 * time.Hour)
	} else {
		// 每周执行
		dayOfWeek := 1 // 默认周一
		if config.DayOfWeek.Valid {
			dayOfWeek = int(config.DayOfWeek.Int64) // 1=周一，7=周日
		}
		
		currentWeekday := int(now.Weekday())
		// Go的Weekday：0=周日，1=周一，...，6=周六
		// 转换为：1=周一，2=周二，...，7=周日
		if currentWeekday == 0 {
			currentWeekday = 7
		}

		// 计算目标日期
		daysUntilTarget := (dayOfWeek - currentWeekday + 7) % 7
		if daysUntilTarget == 0 {
			// 今天就是目标日期
			if today.After(now) || today.Equal(now) {
				// 今天还没到执行时间或正好是执行时间
				return today
			}
			// 今天已经过了执行时间，下周执行
			daysUntilTarget = 7
		}

		targetDate := today.Add(time.Duration(daysUntilTarget) * 24 * time.Hour)
		return targetDate
	}
}

// getScheduleDescription 获取定时配置的描述
func getScheduleDescription(config *ScheduleConfig) string {
	if config.Frequency == "daily" {
		return fmt.Sprintf("每天 %d:00", config.Hour)
	} else {
		dayNames := map[int]string{
			1: "周一",
			2: "周二",
			3: "周三",
			4: "周四",
			5: "周五",
			6: "周六",
			7: "周日",
		}
		dayName := "周一"
		if config.DayOfWeek.Valid {
			if name, ok := dayNames[int(config.DayOfWeek.Int64)]; ok {
				dayName = name
			}
		}
		return fmt.Sprintf("每周%s %d:00", dayName, config.Hour)
	}
}

