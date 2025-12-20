/*
 Navicat Premium Data Transfer

 Source Server         : test
 Source Server Type    : MySQL
 Source Server Version : 50730
 Source Host           : 127.0.0.1:3306
 Source Schema         : ops

 Target Server Type    : MySQL
 Target Server Version : 50730
 File Encoding         : 65001

 Date: 20/12/2025 13:49:48
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for audit_audit_history
-- ----------------------------
DROP TABLE IF EXISTS `audit_audit_history`;
CREATE TABLE `audit_audit_history`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `audit_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '审核意见（历史记录）',
  `audit_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '审核状态（历史记录）',
  `auditor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '审核人用户名',
  `audit_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '审核时间',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '记录创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `task_id`(`task_id`) USING BTREE,
  INDEX `audit_time`(`audit_time`) USING BTREE,
  INDEX `auditor`(`auditor`) USING BTREE,
  CONSTRAINT `audit_audit_history_ibfk_1` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 26 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '设备审核意见历史记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for audit_details
-- ----------------------------
DROP TABLE IF EXISTS `audit_details`;
CREATE TABLE `audit_details`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `device_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '设备编码（*）',
  `original_device_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '原设备编码',
  `device_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '设备名称（*）',
  `division_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '行政区划编码（*）',
  `monitor_point_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '监控点位类型（*）',
  `pickup` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '拾音器',
  `parent_device` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '父设备',
  `construction_unit` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '建设单位/设备归属（*）',
  `construction_unit_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '建设单位/平台归属代码（*）',
  `management_unit` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '管理单位（*）',
  `camera_dept` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '摄像机所属部门（警种）（*）',
  `admin_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '管理员姓名（*）',
  `admin_contact` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '管理员联系电话（*）',
  `contractor` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '承建单位（*）',
  `maintain_unit` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '维护单位（*）',
  `device_vendor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '设备厂商（*）',
  `device_model` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '设备型号',
  `camera_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '摄像机类型（*）',
  `access_method` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '接入方式',
  `camera_function_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '摄像机功能类型（*）',
  `video_encoding_format` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '视频编码格式（*）',
  `image_resolution` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '图像分辨率（*）',
  `camera_light_property` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '摄像机补光属性',
  `backend_structure` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '后端结构化',
  `lens_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '镜头类型',
  `installation_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '安装类型',
  `height_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '高度类型（*）',
  `jurisdiction_police` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '所属辖区公安机关（*）',
  `installation_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '安装地址（*）',
  `surrounding_landmark` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '周边标志（*）',
  `longitude` decimal(10, 6) NOT NULL COMMENT '经度（*）',
  `latitude` decimal(10, 6) NOT NULL COMMENT '纬度（*）',
  `installation_location` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '摄像机安装位置室内外（*）',
  `monitoring_direction` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '摄像机监控方位（*）',
  `pole_number` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '立杆编号（*）',
  `scene_picture` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '摄像机实景图片',
  `networking_property` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '联网属性',
  `access_network` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '接入网络（*）',
  `ipv4_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'IPv4地址（*）',
  `ipv6_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT 'IPv6地址',
  `mac_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '设备MAC地址（*）',
  `access_port` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '访问端口',
  `associated_encoder` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '关联编码器',
  `device_username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '设备用户名',
  `device_password` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '设备口令',
  `channel_number` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '通道号',
  `connection_protocol` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '连接协议',
  `enabled_time` date NULL DEFAULT NULL COMMENT '启用时间（*）',
  `scrapped_time` date NULL DEFAULT NULL COMMENT '报废时间',
  `device_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '设备状态（*）',
  `inspection_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '巡检状态',
  `video_loss` int(11) NULL DEFAULT NULL COMMENT '视频丢失',
  `color_distortion` int(11) NULL DEFAULT NULL COMMENT '色彩失真',
  `video_blur` int(11) NULL DEFAULT NULL COMMENT '视频模糊',
  `brightness_exception` int(11) NULL DEFAULT NULL COMMENT '亮度异常',
  `video_interference` int(11) NULL DEFAULT NULL COMMENT '视频干扰',
  `video_lag` int(11) NULL DEFAULT NULL COMMENT '视频卡顿',
  `video_occlusion` int(11) NULL DEFAULT NULL COMMENT '视频遮挡',
  `scene_change` int(11) NULL DEFAULT NULL COMMENT '场景变更',
  `online_duration` int(11) NULL DEFAULT NULL COMMENT '在线时长',
  `offline_duration` int(11) NULL DEFAULT NULL COMMENT '离线时长',
  `signaling_delay` int(11) NULL DEFAULT NULL COMMENT '信令时延',
  `video_stream_delay` int(11) NULL DEFAULT NULL COMMENT '视频流时延',
  `key_frame_delay` int(11) NULL DEFAULT NULL COMMENT '关键帧时延',
  `recording_retention_days` int(11) NOT NULL COMMENT '录像保存天数（*）',
  `storage_device_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '存储设备编码',
  `storage_channel_number` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '存储通道号',
  `storage_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '存储类型',
  `cache_settings` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '缓存设置',
  `notes` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '备注',
  `collection_area_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '采集区域类型（*）',
  `audit_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '建档状态：0-未审核未建档，1-已审核未建档，2-已建档',
  `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE,
  INDEX `idx_device_code`(`device_code`) USING BTREE,
  INDEX `idx_audit_status`(`audit_status`) USING BTREE,
  UNIQUE INDEX `uk_device_code`(`device_code`) USING BTREE,
  CONSTRAINT `fk_audit_details_task` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1123 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '档案审核明细表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for audit_sample_records
-- ----------------------------
DROP TABLE IF EXISTS `audit_sample_records`;
CREATE TABLE `audit_sample_records`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `sampled_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '抽检人员',
  `sampled_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '抽检时间',
  `sample_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '抽检意见',
  `sample_result` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '抽检结果：通过、待整改',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE,
  INDEX `idx_sampled_at`(`sampled_at`) USING BTREE,
  INDEX `idx_sampled_by`(`sampled_by`) USING BTREE,
  INDEX `idx_sample_result`(`sample_result`) USING BTREE,
  CONSTRAINT `fk_audit_sample_task` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '设备审核抽检记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for audit_tasks
-- ----------------------------
DROP TABLE IF EXISTS `audit_tasks`;
CREATE TABLE `audit_tasks`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '档案名称',
  `organization` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '机构/子公司名称',
  `import_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '导入时间',
  `audit_status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '待审核' COMMENT '审核状态：待审核、已审核待整改、已完成',
  `is_sampled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已抽检：0-未抽检，1-已抽检',
  `last_sampled_at` timestamp(0) NULL DEFAULT NULL COMMENT '最后抽检时间',
  `record_count` int(11) NOT NULL DEFAULT 0 COMMENT '导入记录数量',
  `is_single_soldier` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否单兵设备：0-否，1-是',
  `audit_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '审核意见',
  `auditor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '审核人',
  `audit_time` timestamp(0) NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `updated_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  `archive_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '档案类型：新增、取推、补档案',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_audit_status`(`audit_status`) USING BTREE,
  INDEX `idx_organization`(`organization`) USING BTREE,
  INDEX `idx_import_time`(`import_time`) USING BTREE,
  INDEX `idx_is_sampled`(`is_sampled`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 53 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '档案审核任务表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for audit_video_reminder_schedule_config
-- ----------------------------
DROP TABLE IF EXISTS `audit_video_reminder_schedule_config`;
CREATE TABLE `audit_video_reminder_schedule_config`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `frequency` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'daily' COMMENT '执行频率：daily-每天, weekly-每周',
  `hour` int(11) NOT NULL DEFAULT 1 COMMENT '执行时间（小时）：1-24',
  `day_of_week` int(11) NULL DEFAULT NULL COMMENT '每周执行日期（1-7，1=周一，7=周日），仅当frequency=weekly时有效',
  `enabled` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用：0-禁用，1-启用',
  `updated_at` datetime(0) NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  `updated_by` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '更新人',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_config`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '设备审核录像提醒定时任务配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for audit_video_reminders
-- ----------------------------
DROP TABLE IF EXISTS `audit_video_reminders`;
CREATE TABLE `audit_video_reminders`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '关联的审核任务ID',
  `earliest_video_date` date NOT NULL COMMENT '最早录像日期（从审核意见中提取）',
  `required_days` int(11) NOT NULL COMMENT '要求的天数（30/90/180）',
  `actual_days` int(11) NOT NULL COMMENT '实际天数（审核时计算：审核日期 - 最早录像日期）',
  `reminder_date` date NOT NULL COMMENT '提醒日期（最早录像日期 + 要求天数）',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT 'pending' COMMENT '状态：pending-待处理, notified-已通知, completed-已完成',
  `created_at` datetime(0) NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `notified_at` datetime(0) NULL DEFAULT NULL COMMENT '通知时间',
  `completed_at` datetime(0) NULL DEFAULT NULL COMMENT '完成时间（标记为已处理的时间）',
  `completed_by` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '处理人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_reminder_date`(`reminder_date`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE,
  CONSTRAINT `audit_video_reminders_ibfk_1` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '设备审核录像天数不足提醒表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for checkpoint_audit_history
-- ----------------------------
DROP TABLE IF EXISTS `checkpoint_audit_history`;
CREATE TABLE `checkpoint_audit_history`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联checkpoint_tasks表',
  `audit_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '审核意见（历史记录）',
  `audit_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '审核状态（历史记录）',
  `auditor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '审核人用户名',
  `audit_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '审核时间',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '记录创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `task_id`(`task_id`) USING BTREE,
  INDEX `audit_time`(`audit_time`) USING BTREE,
  INDEX `auditor`(`auditor`) USING BTREE,
  CONSTRAINT `checkpoint_audit_history_ibfk_1` FOREIGN KEY (`task_id`) REFERENCES `checkpoint_tasks` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '卡口审核意见历史记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for checkpoint_details
-- ----------------------------
DROP TABLE IF EXISTS `checkpoint_details`;
CREATE TABLE `checkpoint_details`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联checkpoint_tasks表',
  `checkpoint_code` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口编号（*）',
  `original_checkpoint_code` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '原卡口编号',
  `checkpoint_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口名称（*）',
  `checkpoint_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口地址（*）',
  `road_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '道路名称（*）',
  `direction_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '方向类型',
  `direction_description` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '方向描述（*）',
  `direction_notes` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '方向备注',
  `division_code` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '行政区划（*）',
  `road_section_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '路段类型（*）',
  `road_code` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '道路代码（*）',
  `kilometer_or_intersection_number` varchar(6) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '公里数/路口号（*）',
  `road_meter` varchar(6) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '道路米数（*）',
  `pole_number` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '立杆编号',
  `checkpoint_point_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口点位类型（*）',
  `checkpoint_location_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口位置类型（*）',
  `checkpoint_application_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口应用类型（*）',
  `has_interception_condition` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备拦截条件（*）',
  `has_speed_measurement` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备车辆测速功能',
  `has_realtime_video` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备实时视频功能',
  `has_face_capture` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备人脸抓拍功能',
  `has_violation_capture` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备违章抓拍功能',
  `has_frontend_secondary_recognition` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '具备前端二次识别功能',
  `is_boundary_checkpoint` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '是否边界卡口（*）',
  `adjacent_area` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '邻界地域（*）',
  `checkpoint_longitude` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口经度（*）',
  `checkpoint_latitude` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口纬度（*）',
  `checkpoint_scene_photo_url` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口实景照片地址',
  `checkpoint_status` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口状态（*）',
  `capture_trigger_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '抓拍触发类型',
  `capture_direction_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '抓拍方向类型（*）',
  `total_lanes` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '车道总数（*）',
  `panoramic_camera_device_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '全景球机设备编码（*）',
  `next_checkpoint_along_road` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '沿线下一卡口编号',
  `next_checkpoint_opposite` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '对向下一卡口编号',
  `next_checkpoint_left_turn` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '左转下一卡口编号',
  `next_checkpoint_right_turn` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '右转下一卡口编号',
  `next_checkpoint_u_turn` varchar(18) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '掉头下一卡口编号',
  `construction_unit` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '建设单位（*）',
  `management_unit` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '管理单位（*）',
  `checkpoint_department` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口所属部门（*）',
  `admin_name` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '管理员姓名（*）',
  `admin_contact` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '管理员联系电话',
  `checkpoint_contractor` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口承建单位',
  `checkpoint_maintain_unit` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口维护单位',
  `alarm_receiving_department` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '接警部门（*）',
  `alarm_receiving_department_code` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '接警部门代码（*）',
  `alarm_receiving_phone` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '接警电话（*）',
  `interception_department` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '拦截部门（*）',
  `interception_department_code` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '拦截部门代码（*）',
  `interception_department_contact` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '拦截部门联系电话（*）',
  `terminal_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端编码',
  `terminal_ip_address` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端IP地址（*）',
  `terminal_port` varchar(5) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端端口',
  `terminal_username` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端用户名',
  `terminal_password` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端密码',
  `terminal_vendor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端厂商',
  `checkpoint_enabled_time` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口启用时间（*）',
  `checkpoint_revoked_time` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口撤销时间',
  `notes` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '备注',
  `checkpoint_device_type` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口设备类型（*）',
  `total_capture_cameras` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '抓拍摄像机总数（*）',
  `central_control_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机编码',
  `central_control_ip_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机IP地址',
  `central_control_port` varchar(5) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机端口',
  `central_control_username` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机用户名',
  `central_control_password` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机密码',
  `central_control_vendor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '中控机厂商',
  `checkpoint_scrapped_time` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '卡口报废时间',
  `total_antennas` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '天线总数',
  `terminal_mac_address` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '终端MAC地址（*）',
  `collection_area_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '采集区域类型',
  `integrated_command_platform_checkpoint_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '集成指挥平台卡口编号（组）',
  `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  `audit_status` int(11) NOT NULL DEFAULT 0 COMMENT '建档状态：0-未审核未建档，1-已审核未建档，2-已建档',
  PRIMARY KEY (`id`, `capture_direction_type`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE,
  INDEX `idx_checkpoint_code`(`checkpoint_code`) USING BTREE,
  UNIQUE INDEX `uk_checkpoint_code`(`checkpoint_code`) USING BTREE,
  CONSTRAINT `fk_checkpoint_details_task` FOREIGN KEY (`task_id`) REFERENCES `checkpoint_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 267 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '卡口审核明细表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for checkpoint_sample_records
-- ----------------------------
DROP TABLE IF EXISTS `checkpoint_sample_records`;
CREATE TABLE `checkpoint_sample_records`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '审核任务ID，关联checkpoint_tasks表',
  `sampled_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '抽检人员',
  `sampled_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '抽检时间',
  `sample_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '抽检意见',
  `sample_result` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '抽检结果：通过、待整改',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_task_id`(`task_id`) USING BTREE,
  INDEX `idx_sampled_at`(`sampled_at`) USING BTREE,
  INDEX `idx_sampled_by`(`sampled_by`) USING BTREE,
  INDEX `idx_sample_result`(`sample_result`) USING BTREE,
  CONSTRAINT `fk_checkpoint_sample_task` FOREIGN KEY (`task_id`) REFERENCES `checkpoint_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '卡口审核抽检记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for checkpoint_tasks
-- ----------------------------
DROP TABLE IF EXISTS `checkpoint_tasks`;
CREATE TABLE `checkpoint_tasks`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '档案名称',
  `organization` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '机构/子公司名称',
  `import_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '导入时间',
  `audit_status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '未审核' COMMENT '审核状态：未审核、已审核待整改、已完成',
  `is_sampled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已抽检：0-未抽检，1-已抽检',
  `last_sampled_at` timestamp(0) NULL DEFAULT NULL COMMENT '最后抽检时间',
  `record_count` int(11) NOT NULL DEFAULT 0 COMMENT '导入记录数量',
  `audit_comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '审核意见',
  `auditor` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '审核人',
  `audit_time` timestamp(0) NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `updated_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  `archive_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '档案类型：新增、取推、补档案',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_audit_status`(`audit_status`) USING BTREE,
  INDEX `idx_organization`(`organization`) USING BTREE,
  INDEX `idx_import_time`(`import_time`) USING BTREE,
  INDEX `idx_is_sampled`(`is_sampled`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 23 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '卡口审核任务表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for operation_logs
-- ----------------------------
DROP TABLE IF EXISTS `operation_logs`;
CREATE TABLE `operation_logs`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `action` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `ip` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_created_at`(`created_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1061 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户操作日志' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for organizations
-- ----------------------------
DROP TABLE IF EXISTS `organizations`;
CREATE TABLE `organizations`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '机构名称',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `updated_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_name`(`name`) USING BTREE,
  INDEX `idx_name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '机构名称字典表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for system_settings
-- ----------------------------
DROP TABLE IF EXISTS `system_settings`;
CREATE TABLE `system_settings`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `param_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '参数键名',
  `param_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '参数值',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '参数描述',
  `create_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_param_key`(`param_key`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 78 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '系统参数设置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_role
-- ----------------------------
DROP TABLE IF EXISTS `user_role`;
CREATE TABLE `user_role`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `role_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '角色名称',
  `role_code` tinyint(4) NOT NULL COMMENT '角色代码：0=管理员，1=普通用户',
  `permissions` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '角色权限（JSON格式，预留）',
  `create_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_role_code`(`role_code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户角色表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户名',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码（bcrypt加密）',
  `role_id` int(11) NOT NULL COMMENT '角色ID，关联user_role表',
  `create_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间',
  `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_username`(`username`) USING BTREE,
  INDEX `idx_role_id`(`role_id`) USING BTREE,
  CONSTRAINT `fk_user_role` FOREIGN KEY (`role_id`) REFERENCES `user_role` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
