-- MySQL dump 10.13  Distrib 5.7.30, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: ops
-- ------------------------------------------------------
-- Server version	5.7.30

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `checkpoint_details`
--

DROP TABLE IF EXISTS `checkpoint_details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `checkpoint_details` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) unsigned NOT NULL COMMENT '审核任务ID，关联checkpoint_tasks表',
  `checkpoint_code` varchar(18) DEFAULT NULL COMMENT '卡口编号（*）',
  `original_checkpoint_code` varchar(18) DEFAULT NULL COMMENT '原卡口编号',
  `checkpoint_name` varchar(255) DEFAULT NULL COMMENT '卡口名称（*）',
  `checkpoint_address` varchar(255) DEFAULT NULL COMMENT '卡口地址（*）',
  `road_name` varchar(255) DEFAULT NULL COMMENT '道路名称（*）',
  `direction_type` varchar(50) DEFAULT NULL COMMENT '方向类型',
  `direction_description` varchar(2) DEFAULT NULL COMMENT '方向描述（*）',
  `direction_notes` varchar(255) DEFAULT NULL COMMENT '方向备注',
  `division_code` varchar(8) DEFAULT NULL COMMENT '行政区划（*）',
  `road_section_type` varchar(2) DEFAULT NULL COMMENT '路段类型（*）',
  `road_code` varchar(8) DEFAULT NULL COMMENT '道路代码（*）',
  `kilometer_or_intersection_number` varchar(6) DEFAULT NULL COMMENT '公里数/路口号（*）',
  `road_meter` varchar(6) DEFAULT NULL COMMENT '道路米数（*）',
  `pole_number` varchar(30) DEFAULT NULL COMMENT '立杆编号',
  `checkpoint_point_type` varchar(2) DEFAULT NULL COMMENT '卡口点位类型（*）',
  `checkpoint_location_type` varchar(2) DEFAULT NULL COMMENT '卡口位置类型（*）',
  `checkpoint_application_type` varchar(2) DEFAULT NULL COMMENT '卡口应用类型（*）',
  `has_interception_condition` varchar(2) DEFAULT NULL COMMENT '具备拦截条件（*）',
  `has_speed_measurement` varchar(2) DEFAULT NULL COMMENT '具备车辆测速功能',
  `has_realtime_video` varchar(2) DEFAULT NULL COMMENT '具备实时视频功能',
  `has_face_capture` varchar(2) DEFAULT NULL COMMENT '具备人脸抓拍功能',
  `has_violation_capture` varchar(2) DEFAULT NULL COMMENT '具备违章抓拍功能',
  `has_frontend_secondary_recognition` varchar(2) DEFAULT NULL COMMENT '具备前端二次识别功能',
  `is_boundary_checkpoint` varchar(2) DEFAULT NULL COMMENT '是否边界卡口（*）',
  `adjacent_area` varchar(20) DEFAULT NULL COMMENT '邻界地域（*）',
  `checkpoint_longitude` varchar(15) DEFAULT NULL COMMENT '卡口经度（*）',
  `checkpoint_latitude` varchar(15) DEFAULT NULL COMMENT '卡口纬度（*）',
  `checkpoint_scene_photo_url` varchar(30) DEFAULT NULL COMMENT '卡口实景照片地址',
  `checkpoint_status` varchar(2) DEFAULT NULL COMMENT '卡口状态（*）',
  `capture_trigger_type` varchar(2) DEFAULT NULL COMMENT '抓拍触发类型',
  `capture_direction_type` varchar(2) NOT NULL COMMENT '抓拍方向类型（*）',
  `total_lanes` varchar(2) DEFAULT NULL COMMENT '车道总数（*）',
  `panoramic_camera_device_code` varchar(20) DEFAULT NULL COMMENT '全景球机设备编码（*）',
  `next_checkpoint_along_road` varchar(18) DEFAULT NULL COMMENT '沿线下一卡口编号',
  `next_checkpoint_opposite` varchar(18) DEFAULT NULL COMMENT '对向下一卡口编号',
  `next_checkpoint_left_turn` varchar(18) DEFAULT NULL COMMENT '左转下一卡口编号',
  `next_checkpoint_right_turn` varchar(18) DEFAULT NULL COMMENT '右转下一卡口编号',
  `next_checkpoint_u_turn` varchar(18) DEFAULT NULL COMMENT '掉头下一卡口编号',
  `construction_unit` varchar(30) DEFAULT NULL COMMENT '建设单位（*）',
  `management_unit` varchar(30) DEFAULT NULL COMMENT '管理单位（*）',
  `checkpoint_department` varchar(2) DEFAULT NULL COMMENT '卡口所属部门（*）',
  `admin_name` varchar(8) DEFAULT NULL COMMENT '管理员姓名（*）',
  `admin_contact` varchar(15) DEFAULT NULL COMMENT '管理员联系电话',
  `checkpoint_contractor` varchar(30) DEFAULT NULL COMMENT '卡口承建单位',
  `checkpoint_maintain_unit` varchar(50) DEFAULT NULL COMMENT '卡口维护单位',
  `alarm_receiving_department` varchar(15) DEFAULT NULL COMMENT '接警部门（*）',
  `alarm_receiving_department_code` varchar(15) DEFAULT NULL COMMENT '接警部门代码（*）',
  `alarm_receiving_phone` varchar(15) DEFAULT NULL COMMENT '接警电话（*）',
  `interception_department` varchar(15) DEFAULT NULL COMMENT '拦截部门（*）',
  `interception_department_code` varchar(15) DEFAULT NULL COMMENT '拦截部门代码（*）',
  `interception_department_contact` varchar(15) DEFAULT NULL COMMENT '拦截部门联系电话（*）',
  `terminal_code` varchar(20) DEFAULT NULL COMMENT '终端编码',
  `terminal_ip_address` varchar(20) DEFAULT NULL COMMENT '终端IP地址（*）',
  `terminal_port` varchar(5) DEFAULT NULL COMMENT '终端端口',
  `terminal_username` varchar(10) DEFAULT NULL COMMENT '终端用户名',
  `terminal_password` varchar(20) DEFAULT NULL COMMENT '终端密码',
  `terminal_vendor` varchar(50) DEFAULT NULL COMMENT '终端厂商',
  `checkpoint_enabled_time` varchar(30) DEFAULT NULL COMMENT '卡口启用时间（*）',
  `checkpoint_revoked_time` varchar(30) DEFAULT NULL COMMENT '卡口撤销时间',
  `notes` varchar(20) DEFAULT NULL COMMENT '备注',
  `checkpoint_device_type` varchar(2) DEFAULT NULL COMMENT '卡口设备类型（*）',
  `total_capture_cameras` varchar(2) DEFAULT NULL COMMENT '抓拍摄像机总数（*）',
  `central_control_code` varchar(20) DEFAULT NULL COMMENT '中控机编码',
  `central_control_ip_address` varchar(50) DEFAULT NULL COMMENT '中控机IP地址',
  `central_control_port` varchar(5) DEFAULT NULL COMMENT '中控机端口',
  `central_control_username` varchar(20) DEFAULT NULL COMMENT '中控机用户名',
  `central_control_password` varchar(20) DEFAULT NULL COMMENT '中控机密码',
  `central_control_vendor` varchar(50) DEFAULT NULL COMMENT '中控机厂商',
  `checkpoint_scrapped_time` varchar(20) DEFAULT NULL COMMENT '卡口报废时间',
  `total_antennas` varchar(20) DEFAULT NULL COMMENT '天线总数',
  `terminal_mac_address` varchar(30) DEFAULT NULL COMMENT '终端MAC地址（*）',
  `collection_area_type` varchar(20) DEFAULT NULL COMMENT '采集区域类型',
  `integrated_command_platform_checkpoint_code` varchar(20) DEFAULT NULL COMMENT '集成指挥平台卡口编号（组）',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `audit_status` int(11) NOT NULL DEFAULT '0' COMMENT '建档状态：0-未审核未建档，1-已审核未建档，2-已建档',
  PRIMARY KEY (`id`,`capture_direction_type`) USING BTREE,
  KEY `idx_task_id` (`task_id`),
  KEY `idx_checkpoint_code` (`checkpoint_code`),
  CONSTRAINT `fk_checkpoint_details_task` FOREIGN KEY (`task_id`) REFERENCES `checkpoint_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=123 DEFAULT CHARSET=utf8mb4 COMMENT='卡口审核明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `checkpoint_details`
--

LOCK TABLES `checkpoint_details` WRITE;
/*!40000 ALTER TABLE `checkpoint_details` DISABLE KEYS */;
INSERT INTO `checkpoint_details` VALUES (103,14,'442016100510700109',NULL,'中山市三角镇沙栏西路-福源南路路口东行','中山市三角镇沙栏西路-福源南路路口东行','中山市三角镇沙栏西路-福源南路路口东行',NULL,'1',NULL,'442016','9','94729','1023','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.41056','22.65842',NULL,'3','3','2','1','44201610051310700122',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(104,14,'442016100510700110',NULL,'中山市三角镇沙栏西路-福源南路路口西行','中山市三角镇沙栏西路-福源南路路口西行','中山市三角镇沙栏西路-福源南路路口西行',NULL,'2',NULL,'442016','9','94729','1025','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.41147','22.65797',NULL,'3','3','2','1','44201610051310700121',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(105,14,'442016100510700111',NULL,'中山市三角镇福源南路-沙栏西路路口南行\n车道1，2','中山市三角镇福源南路-沙栏西路路口南行\n车道1，2','中山市三角镇福源南路-沙栏西路路口南行\n车道1，2',NULL,'3',NULL,'442016','9','94729','1021','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.411219','22.658776',NULL,'3','3','2','2','44201610051310700124',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(106,14,'442016100510700112',NULL,'中山市三角镇福源南路-沙栏西路路口南行 \n车道3，4','中山市三角镇福源南路-沙栏西路路口南行 \n车道3，4','中山市三角镇福源南路-沙栏西路路口南行 \n车道3，4',NULL,'3',NULL,'442016','9','94729','1021','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.411219','22.658776',NULL,'3','3','2','2','44201610051310700124',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(107,14,'442016100510700113',NULL,'中山市三角镇福源南路-沙栏西路路口北行\n车道1，2','中山市三角镇福源南路-沙栏西路路口北行\n车道1，2','中山市三角镇福源南路-沙栏西路路口北行\n车道1，2',NULL,'4',NULL,'442016','9','94702','1005','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.41094','22.65784',NULL,'3','3','2','2','44201610051310700123',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(108,14,'442016100510700114',NULL,'中山市三角镇福源南路-沙栏西路路口北行\n车道3，4','中山市三角镇福源南路-沙栏西路路口北行\n车道3，4','中山市三角镇福源南路-沙栏西路路口北行\n车道3，4',NULL,'4',NULL,'442016','9','94729','1021','150',NULL,'1','4','2','0','0','1','0','1','0','0','无','113.41094','22.65784',NULL,'3','3','2','2','44201610051310700123',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(109,14,'442016100510700115',NULL,'中山市三角镇沙栏西路-福源南路路口东行（反向抓拍）','中山市三角镇沙栏西路-福源南路路口东行（反向抓拍）','中山市三角镇沙栏西路-福源南路路口东行（反向抓拍）',NULL,'1',NULL,'442016','9','94729','1023','150',NULL,'1','4','3','0','0','1','1','1','0','0','无','113.41056','22.65842',NULL,'3','3','1','1','44201610051310700122',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(110,14,'442016100510700116',NULL,'中山市三角镇沙栏西路-福源南路路口西行（反向抓拍）','中山市三角镇沙栏西路-福源南路路口西行（反向抓拍）','中山市三角镇沙栏西路-福源南路路口西行（反向抓拍）',NULL,'2',NULL,'442016','9','94729','1020','150',NULL,'1','4','2','0','0','1','1','1','0','0','无','113.41147','22.65797',NULL,'3','3','1','1','44201610051310700121',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(111,14,'442016100510700117',NULL,'中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道1，2','中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道1，2','中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道1，2',NULL,'3',NULL,'442016','9','94729','1015','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.411219','22.658776',NULL,'3','2','1','2','44201610051310700124',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(112,14,'442016100510700118',NULL,'中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道3，4','中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道3，4','中山市三角镇福源南路-沙栏西路路口南行 \n（反向抓拍）车道3，4',NULL,'3',NULL,'442016','9','94729','1021','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.411219','22.658776',NULL,'3','2','1','2','44201610051310700124',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(113,14,'442016100510700119',NULL,'中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道1，2','中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道1，2','中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道1，2',NULL,'4',NULL,'442016','9','94731','1003','150',NULL,'1','4','3','0','0','1','1','1','0','0','无','113.41094','22.65784',NULL,'3','3','1','2','44201610051310700123',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(114,14,'442016100510700120',NULL,'中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道3，4','中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道3，4','中山市三角镇福源南路-沙栏西路路口北行 \n（反向抓拍）车道3，4',NULL,'4',NULL,'442016','9','94729','1021','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.41094','22.65784',NULL,'3','3','1','2','44201610051310700123',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,'44.103.48.185',NULL,'admin','Sjfj@23185005','1','2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'e8:a0:ed:28:90:8b',NULL,NULL,'2025-12-18 01:40:24',0),(115,14,'442016100510700130',NULL,'中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道3','中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道3','中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道3',NULL,'1',NULL,'442016','9','94729','1019','150',NULL,'1','4','2','0','0','1','1','1','0','0','无','113.414074','22.671368',NULL,'3','3','1','1','44201610051310700134',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:44:2e',NULL,NULL,'2025-12-18 01:40:24',0),(116,14,'442016100510700131',NULL,'中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道1，2','中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道1，2','中山市三角镇和谐路-福源南路路口东行（反向抓拍）车道1，2',NULL,'1',NULL,'442016','9','94729','1021','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.414074','22.671368',NULL,'3','3','1','2','44201610051310700134',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:7c:6e:63',NULL,NULL,'2025-12-18 01:40:24',0),(117,14,'442016100510700132',NULL,'中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道1','中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道1','中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道1',NULL,'2',NULL,'442016','9','94702','1005','150',NULL,'1','4','3','0','0','1','1','1','0','0','无','113.414793','22.671467',NULL,'3','3','1','1','44201610051310700133',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:45:da',NULL,NULL,'2025-12-18 01:40:24',0),(118,14,'442016100510700133',NULL,'中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道2，3','中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道2，3','中山市三角镇孝福路-福源南路路口西行（反向抓拍）车道2，3',NULL,'2',NULL,'442016','9','94729','1021','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.414793','22.671467',NULL,'3','3','1','2','44201610051310700133',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:44:30',NULL,NULL,'2025-12-18 01:40:24',0),(119,14,'442016100510700134',NULL,'中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道1，2','中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道1，2','中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道1，2',NULL,'3',NULL,'442016','9','94729','1024','150',NULL,'1','4','2','0','0','1','1','1','0','0','无','113.414385','22.671759',NULL,'3','3','1','2','44201610051310700136',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:45:de',NULL,NULL,'2025-12-18 01:40:24',0),(120,14,'442016100510700135',NULL,'中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道3，4','中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道3，4','中山市三角镇福源南路-和谐路路口南行（反向抓拍）车道3，4',NULL,'3',NULL,'442016','9','94729','1021','150',NULL,'1','4','3','0','1','1','1','1','0','0','无','113.414385','22.671759',NULL,'3','3','1','2','44201610051310700136',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:7c:6e:7f',NULL,NULL,'2025-12-18 01:40:24',0),(121,14,'442016100510700136',NULL,'中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道1，2','中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道1，2','中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道1，2',NULL,'4',NULL,'442016','9','94729','1019','150',NULL,'1','4','3','0','0','1','1','1','0','0','无','113.414573','22.671056',NULL,'3','3','1','2','44201610051310700135',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:45:b0',NULL,NULL,'2025-12-18 01:40:24',0),(122,14,'442016100510700137',NULL,'中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道3，4','中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道3，4','中山市三角镇福源南路-和谐路路口北行（反向抓拍）车道3，4',NULL,'4',NULL,'442016','9','94729','1019','150',NULL,'1','4','3','0','0','1','1','1','0','0','无','113.414573','22.671056',NULL,'3','3','1','2','44201610051310700135',NULL,NULL,NULL,NULL,NULL,'中山市公安局三角分局','中山市公安局三角交警大队','11',NULL,NULL,'广东省广播电视网络股份有限公司','广东省广播电视网络股份有限公司-联系人-黄永强-电话-18928119877',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'admin','Sjfj@23185005',NULL,'2024-12-25 00:00:00.0',NULL,NULL,'1','1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'24:48:45:80:45:a0',NULL,NULL,'2025-12-18 01:40:24',0);
/*!40000 ALTER TABLE `checkpoint_details` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-12-18 11:17:15
