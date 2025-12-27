package overview

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"ops-web/internal/db"
	"ops-web/internal/logger"
)

// 统计数据
type OverviewStats struct {
	Total           int            `json:"total"`            // 导入档案总数
	Pending         int            `json:"pending"`           // 未审核
	NeedFix         int            `json:"need_fix"`          // 已审核待整改
	Completed       int            `json:"completed"`         // 已完成
	PendingSample   int            `json:"pending_sample"`     // 待抽检
	SampleNeedFix   int            `json:"sample_need_fix"`   // 抽检后待整改
	SamplePassed    int            `json:"sample_passed"`     // 抽检已通过
	TaggedStats     map[string]int `json:"tagged_stats"`     // 已打标签档案数量（按tag值分组）
}

// 页面数据
type OverviewPageData struct {
	Title      string
	ActiveMenu string
	SubMenu    string
	TimeRange  string // 时间段：week, month, all
}

// 主入口 - 数据概览页面
func Handler(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "week" // 默认最近一周
	}

	data := OverviewPageData{
		Title:      "数据概览",
		ActiveMenu: "overview",
		SubMenu:    "",
		TimeRange:  timeRange,
	}

	tmpl, err := template.ParseFiles("templates/overview.html")
	if err != nil {
		logger.Errorf("模板解析失败: %v", err)
		http.Error(w, "模板解析失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("模板渲染失败: %v", err)
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// API接口 - 获取设备统计数据
func DeviceStatsAPI(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "week" // 默认最近一周
	}

	// 添加调试日志
	logger.Infof("数据概览API-收到请求，时间范围: %s", timeRange)

	// 设置响应头，禁用缓存
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	stats, err := getDeviceStats(timeRange)
	if err != nil {
		logger.Errorf("获取设备统计数据失败: %v", err)
		http.Error(w, "获取统计数据失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// 获取设备统计数据
// 统计：导入档案总数、未审核、已审核待整改、已完成、待抽检、抽检后待整改、抽检已通过
// 未审核和已审核待整改基于import_time，已完成基于completed_at
// 待抽检基于completed_at，抽检后待整改和抽检已通过基于last_sampled_at
func getDeviceStats(timeRange string) (*OverviewStats, error) {
	// 构建时间条件
	// 对于未审核和已审核待整改，使用import_time
	// 对于已完成，使用completed_at
	// 对于待抽检，使用completed_at
	// 对于抽检后待整改和抽检已通过，使用last_sampled_at
	var timeCondition string
	var args []interface{}

	switch timeRange {
	case "week":
		// 时间筛选逻辑：
		// 1. 未审核和已审核待整改：基于import_time
		// 2. 已完成且未抽检（待抽检）：基于completed_at
		// 3. 已完成且已抽检（抽检后待整改、抽检已通过）：基于last_sampled_at
		timeCondition = `AND (
			(at.audit_status IN ('未审核', '已审核待整改') AND at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY))
			OR 
			(at.audit_status = '已完成' AND (at.is_sampled = 0 OR at.is_sampled IS NULL) AND at.completed_at >= DATE_SUB(NOW(), INTERVAL 7 DAY))
			OR
			(at.audit_status = '已完成' AND at.is_sampled = 1 AND at.last_sampled_at >= DATE_SUB(NOW(), INTERVAL 7 DAY))
		)`
	case "month":
		timeCondition = `AND (
			(at.audit_status IN ('未审核', '已审核待整改') AND at.import_time >= DATE_SUB(NOW(), INTERVAL 1 MONTH))
			OR 
			(at.audit_status = '已完成' AND (at.is_sampled = 0 OR at.is_sampled IS NULL) AND at.completed_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH))
			OR
			(at.audit_status = '已完成' AND at.is_sampled = 1 AND at.last_sampled_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH))
		)`
	case "all":
		timeCondition = ""
	default:
		timeCondition = `AND (
			(at.audit_status IN ('未审核', '已审核待整改') AND at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY))
			OR 
			(at.audit_status = '已完成' AND (at.is_sampled = 0 OR at.is_sampled IS NULL) AND at.completed_at >= DATE_SUB(NOW(), INTERVAL 7 DAY))
			OR
			(at.audit_status = '已完成' AND at.is_sampled = 1 AND at.last_sampled_at >= DATE_SUB(NOW(), INTERVAL 7 DAY))
		)`
	}

	// 统计七个指标：导入档案总数、未审核、已审核待整改、已完成、待抽检、抽检后待整改、抽检已通过
	// 需要关联audit_sample_records表获取最近一次抽检结果
	// 使用COALESCE处理SUM可能返回NULL的情况
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN at.audit_status = '未审核' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN at.audit_status = '已审核待整改' THEN 1 ELSE 0 END), 0) as need_fix,
			COALESCE(SUM(CASE WHEN at.audit_status = '已完成' THEN 1 ELSE 0 END), 0) as completed,
			-- 待抽检：已完成且未抽检
			COALESCE(SUM(CASE WHEN at.audit_status = '已完成' AND (at.is_sampled = 0 OR at.is_sampled IS NULL) THEN 1 ELSE 0 END), 0) as pending_sample,
			-- 抽检后待整改：已完成且已抽检，且最近一次抽检结果为"待整改"
			COALESCE(SUM(CASE WHEN at.audit_status = '已完成' 
				AND at.is_sampled = 1 
				AND latest_sample.sample_result = '待整改' 
				THEN 1 ELSE 0 END), 0) as sample_need_fix,
			-- 抽检已通过：已完成且已抽检，且最近一次抽检结果为"通过"或NULL
			COALESCE(SUM(CASE WHEN at.audit_status = '已完成' 
				AND at.is_sampled = 1 
				AND (latest_sample.sample_result = '通过' OR latest_sample.sample_result IS NULL)
				THEN 1 ELSE 0 END), 0) as sample_passed
		FROM audit_tasks at
		LEFT JOIN (
			SELECT 
				asr1.task_id,
				asr1.sample_result
			FROM audit_sample_records asr1
			INNER JOIN (
				SELECT task_id, MAX(sampled_at) as max_sampled_at
				FROM audit_sample_records
				GROUP BY task_id
			) asr2 ON asr1.task_id = asr2.task_id AND asr1.sampled_at = asr2.max_sampled_at
		) latest_sample ON at.id = latest_sample.task_id
		WHERE 1=1 %s
	`, timeCondition)

	// 添加调试日志
	logger.Infof("数据概览-查询SQL: %s, 参数: %v, 时间范围: %s", query, args, timeRange)

	var stats OverviewStats
	err := db.DBInstance.QueryRow(query, args...).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.NeedFix,
		&stats.Completed,
		&stats.PendingSample,
		&stats.SampleNeedFix,
		&stats.SamplePassed,
	)

	if err != nil {
		logger.Errorf("数据概览-查询失败: %v, SQL: %s", err, query)
		return nil, fmt.Errorf("查询设备统计数据失败: %v", err)
	}

	// 获取tag分组统计
	taggedStats, err := getTaggedStats(timeRange)
	if err != nil {
		logger.Errorf("数据概览-获取tag统计失败: %v", err)
		stats.TaggedStats = make(map[string]int)
	} else {
		stats.TaggedStats = taggedStats
	}

	// 添加调试日志
	logger.Infof("数据概览-查询结果: Total=%d, Pending=%d, NeedFix=%d, Completed=%d, PendingSample=%d, SampleNeedFix=%d, SamplePassed=%d, TaggedStats=%v", 
		stats.Total, stats.Pending, stats.NeedFix, stats.Completed, stats.PendingSample, stats.SampleNeedFix, stats.SamplePassed, stats.TaggedStats)

	return &stats, nil
}

// getTaggedStats: 获取tag分组统计
func getTaggedStats(timeRange string) (map[string]int, error) {
	// 构建tag统计的时间条件（使用import_time）
	var tagTimeCondition string
	switch timeRange {
	case "week":
		tagTimeCondition = "AND at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)"
	case "month":
		tagTimeCondition = "AND at.import_time >= DATE_SUB(NOW(), INTERVAL 1 MONTH)"
	case "all":
		tagTimeCondition = ""
	default:
		tagTimeCondition = "AND at.import_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)"
	}

	// 查询tag分组统计
	query := fmt.Sprintf(`
		SELECT 
			at.tag,
			COUNT(*) as count
		FROM audit_tasks at
		WHERE 1=1 
		AND at.tag IS NOT NULL 
		AND at.tag != ''
		%s
		GROUP BY at.tag
		ORDER BY at.tag
	`, tagTimeCondition)

	logger.Infof("数据概览-tag统计SQL: %s, 时间范围: %s", query, timeRange)

	rows, err := db.DBInstance.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询tag统计失败: %v", err)
	}
	defer rows.Close()

	taggedStats := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			logger.Errorf("数据概览-扫描tag统计失败: %v", err)
			continue
		}
		taggedStats[tag] = count
	}

	return taggedStats, nil
}

// API接口 - 获取卡口统计数据（预留）
func CheckpointStatsAPI(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "week"
	}

	stats, err := getCheckpointStats(timeRange)
	if err != nil {
		logger.Errorf("获取卡口统计数据失败: %v", err)
		http.Error(w, "获取统计数据失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(stats)
}

// 获取卡口统计数据（预留，待实现）
func getCheckpointStats(timeRange string) (*OverviewStats, error) {
	// TODO: 实现卡口统计逻辑
	// 暂时返回空数据
	return &OverviewStats{}, nil
}

