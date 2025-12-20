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
-- Table structure for table `audit_details`
--

DROP TABLE IF EXISTS `audit_details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `audit_details` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `task_id` bigint(20) unsigned NOT NULL COMMENT '审核任务ID，关联audit_tasks表',
  `device_code` varchar(20) NOT NULL COMMENT '设备编码（*）',
  `original_device_code` varchar(20) DEFAULT NULL COMMENT '原设备编码',
  `device_name` varchar(50) NOT NULL COMMENT '设备名称（*）',
  `division_code` varchar(50) NOT NULL COMMENT '行政区划编码（*）',
  `monitor_point_type` varchar(50) NOT NULL COMMENT '监控点位类型（*）',
  `pickup` varchar(50) DEFAULT NULL COMMENT '拾音器',
  `parent_device` varchar(50) DEFAULT NULL COMMENT '父设备',
  `construction_unit` varchar(50) NOT NULL COMMENT '建设单位/设备归属（*）',
  `construction_unit_code` varchar(50) NOT NULL COMMENT '建设单位/平台归属代码（*）',
  `management_unit` varchar(50) NOT NULL COMMENT '管理单位（*）',
  `camera_dept` varchar(50) NOT NULL COMMENT '摄像机所属部门（警种）（*）',
  `admin_name` varchar(50) NOT NULL COMMENT '管理员姓名（*）',
  `admin_contact` varchar(50) NOT NULL COMMENT '管理员联系电话（*）',
  `contractor` varchar(100) NOT NULL COMMENT '承建单位（*）',
  `maintain_unit` varchar(100) NOT NULL COMMENT '维护单位（*）',
  `device_vendor` varchar(50) NOT NULL COMMENT '设备厂商（*）',
  `device_model` varchar(50) DEFAULT NULL COMMENT '设备型号',
  `camera_type` varchar(50) NOT NULL COMMENT '摄像机类型（*）',
  `access_method` varchar(50) DEFAULT NULL COMMENT '接入方式',
  `camera_function_type` varchar(50) NOT NULL COMMENT '摄像机功能类型（*）',
  `video_encoding_format` varchar(50) NOT NULL COMMENT '视频编码格式（*）',
  `image_resolution` varchar(50) NOT NULL COMMENT '图像分辨率（*）',
  `camera_light_property` varchar(50) DEFAULT NULL COMMENT '摄像机补光属性',
  `backend_structure` varchar(50) DEFAULT NULL COMMENT '后端结构化',
  `lens_type` varchar(50) DEFAULT NULL COMMENT '镜头类型',
  `installation_type` varchar(50) DEFAULT NULL COMMENT '安装类型',
  `height_type` varchar(50) NOT NULL COMMENT '高度类型（*）',
  `jurisdiction_police` varchar(50) NOT NULL COMMENT '所属辖区公安机关（*）',
  `installation_address` varchar(50) NOT NULL COMMENT '安装地址（*）',
  `surrounding_landmark` varchar(50) NOT NULL COMMENT '周边标志（*）',
  `longitude` decimal(10,6) NOT NULL COMMENT '经度（*）',
  `latitude` decimal(10,6) NOT NULL COMMENT '纬度（*）',
  `installation_location` varchar(50) NOT NULL COMMENT '摄像机安装位置室内外（*）',
  `monitoring_direction` varchar(50) NOT NULL COMMENT '摄像机监控方位（*）',
  `pole_number` varchar(50) NOT NULL COMMENT '立杆编号（*）',
  `scene_picture` varchar(50) DEFAULT NULL COMMENT '摄像机实景图片',
  `networking_property` varchar(50) DEFAULT NULL COMMENT '联网属性',
  `access_network` varchar(50) NOT NULL COMMENT '接入网络（*）',
  `ipv4_address` varchar(50) NOT NULL COMMENT 'IPv4地址（*）',
  `ipv6_address` varchar(50) DEFAULT NULL COMMENT 'IPv6地址',
  `mac_address` varchar(50) NOT NULL COMMENT '设备MAC地址（*）',
  `access_port` varchar(50) DEFAULT NULL COMMENT '访问端口',
  `associated_encoder` varchar(50) DEFAULT NULL COMMENT '关联编码器',
  `device_username` varchar(50) DEFAULT NULL COMMENT '设备用户名',
  `device_password` varchar(50) DEFAULT NULL COMMENT '设备口令',
  `channel_number` varchar(50) DEFAULT NULL COMMENT '通道号',
  `connection_protocol` varchar(50) DEFAULT NULL COMMENT '连接协议',
  `enabled_time` date DEFAULT NULL COMMENT '启用时间（*）',
  `scrapped_time` date DEFAULT NULL COMMENT '报废时间',
  `device_status` varchar(50) NOT NULL COMMENT '设备状态（*）',
  `inspection_status` varchar(50) DEFAULT NULL COMMENT '巡检状态',
  `video_loss` int(11) DEFAULT NULL COMMENT '视频丢失',
  `color_distortion` int(11) DEFAULT NULL COMMENT '色彩失真',
  `video_blur` int(11) DEFAULT NULL COMMENT '视频模糊',
  `brightness_exception` int(11) DEFAULT NULL COMMENT '亮度异常',
  `video_interference` int(11) DEFAULT NULL COMMENT '视频干扰',
  `video_lag` int(11) DEFAULT NULL COMMENT '视频卡顿',
  `video_occlusion` int(11) DEFAULT NULL COMMENT '视频遮挡',
  `scene_change` int(11) DEFAULT NULL COMMENT '场景变更',
  `online_duration` int(11) DEFAULT NULL COMMENT '在线时长',
  `offline_duration` int(11) DEFAULT NULL COMMENT '离线时长',
  `signaling_delay` int(11) DEFAULT NULL COMMENT '信令时延',
  `video_stream_delay` int(11) DEFAULT NULL COMMENT '视频流时延',
  `key_frame_delay` int(11) DEFAULT NULL COMMENT '关键帧时延',
  `recording_retention_days` int(11) NOT NULL COMMENT '录像保存天数（*）',
  `storage_device_code` varchar(50) DEFAULT NULL COMMENT '存储设备编码',
  `storage_channel_number` varchar(50) DEFAULT NULL COMMENT '存储通道号',
  `storage_type` varchar(50) DEFAULT NULL COMMENT '存储类型',
  `cache_settings` varchar(50) DEFAULT NULL COMMENT '缓存设置',
  `notes` varchar(255) DEFAULT NULL COMMENT '备注',
  `collection_area_type` varchar(50) NOT NULL COMMENT '采集区域类型（*）',
  `audit_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '建档状态：0-未审核未建档，1-已审核未建档，2-已建档',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_device_code` (`device_code`),
  KEY `idx_audit_status` (`audit_status`),
  CONSTRAINT `fk_audit_details_task` FOREIGN KEY (`task_id`) REFERENCES `audit_tasks` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=213 DEFAULT CHARSET=utf8mb4 COMMENT='档案审核明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `audit_details`
--

LOCK TABLES `audit_details` WRITE;
/*!40000 ALTER TABLE `audit_details` DISABLE KEYS */;
INSERT INTO `audit_details` (`id`, `task_id`, `device_code`, `original_device_code`, `device_name`, `division_code`, `monitor_point_type`, `pickup`, `parent_device`, `construction_unit`, `construction_unit_code`, `management_unit`, `camera_dept`, `admin_name`, `admin_contact`, `contractor`, `maintain_unit`, `device_vendor`, `device_model`, `camera_type`, `access_method`, `camera_function_type`, `video_encoding_format`, `image_resolution`, `camera_light_property`, `backend_structure`, `lens_type`, `installation_type`, `height_type`, `jurisdiction_police`, `installation_address`, `surrounding_landmark`, `longitude`, `latitude`, `installation_location`, `monitoring_direction`, `pole_number`, `scene_picture`, `networking_property`, `access_network`, `ipv4_address`, `ipv6_address`, `mac_address`, `access_port`, `associated_encoder`, `device_username`, `device_password`, `channel_number`, `connection_protocol`, `enabled_time`, `scrapped_time`, `device_status`, `inspection_status`, `video_loss`, `color_distortion`, `video_blur`, `brightness_exception`, `video_interference`, `video_lag`, `video_occlusion`, `scene_change`, `online_duration`, `offline_duration`, `signaling_delay`, `video_stream_delay`, `key_frame_delay`, `recording_retention_days`, `storage_device_code`, `storage_channel_number`, `storage_type`, `cache_settings`, `notes`, `collection_area_type`, `audit_status`, `update_time`) VALUES (173,29,'44201701001310004232',NULL,'广东省中山市横栏镇4232纵四线与大涌交界-往大涌方向1,2车道','442017','1','0',NULL,'广东省中山市横栏分局','1','中山市公安局横栏分局','2','郭鸣竹','13928161191','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442017010001','广东省中山市横栏镇4232纵四线与大涌交界-往大涌方向1,2车道','广东省中山市横栏镇4232纵四线与大涌交界-往大涌方向1,2车道',113.263604,22.510164,'1','6','',NULL,'0','2','44.103.12.79',NULL,'80:7c:62:a7:f7:b3',NULL,NULL,'admin','hik-12345',NULL,NULL,'2022-12-01',NULL,'1','1',0,0,0,0,0,0,0,0,0,0,0,0,0,30,NULL,NULL,'1','0',NULL,'A0307',0,'2025-12-18 02:12:21'),(174,29,'44201701001310004233',NULL,'广东省中山市横栏镇4233纵四线与大涌交界-往大涌方向3,4车道','442017','1','0',NULL,'广东省中山市横栏分局','1','中山市公安局横栏分局','2','郭鸣竹','13928161191','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442017010001','广东省中山市横栏镇4233纵四线与大涌交界-往大涌方向3,4车道','广东省中山市横栏镇4233纵四线与大涌交界-往大涌方向3,4车道',113.263604,22.510164,'1','1','',NULL,'0','2','44.103.12.80',NULL,'80:7c:62:a7:f7:a3',NULL,NULL,'admin','hik-12345',NULL,NULL,'2022-12-01',NULL,'1','1',0,0,0,0,0,0,0,0,0,0,0,0,0,30,NULL,NULL,'1','0',NULL,'A0307',0,'2025-12-18 02:12:21'),(175,29,'44201701001310004226',NULL,'广东省中山市横栏镇4226沙古公路横栏与沙溪交界-往沙溪方向2,3车道','442017','1','0',NULL,'广东省中山市横栏分局','1','中山市公安局横栏分局','2','郭鸣竹','13928161191','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,NULL,'1','2','442017010001','广东省中山市横栏镇4226沙古公路横栏与沙溪交界-往沙溪方向2,3车道','广东省中山市横栏镇4226沙古公路横栏与沙溪交界-往沙溪方向2,3车道',113.269997,22.546286,'1','4','',NULL,NULL,'2','44.103.12.73',NULL,'24:32:ae:ee:8c:33',NULL,NULL,'admin','hik-12345',NULL,NULL,'2022-12-01',NULL,'4','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'A0307',0,'2025-12-18 02:12:21'),(176,29,'44201701001310004222',NULL,'广东省中山市横栏镇4222沙古公路横栏与沙溪交界-往横栏方向辅道,1车道','442017','1','0',NULL,'广东省中山市横栏分局','1','中山市公安局横栏分局','2','郭鸣竹','13928161191','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,NULL,'1','2','442017010001','广东省中山市横栏镇4222沙古公路横栏与沙溪交界-往横栏方向辅道,1车道','广东省中山市横栏镇4222沙古公路横栏与沙溪交界-往横栏方向辅道,1车道',113.269997,22.546286,'1','6','',NULL,NULL,'2','44.103.12.69',NULL,'24:32:ae:ee:8b:ed',NULL,NULL,'admin','hik-12345',NULL,NULL,'2022-12-01',NULL,'4','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'A0307',0,'2025-12-18 02:12:21'),(177,30,'44201400071217960113',NULL,'中山坦洲镇大兴酒店停车场出口','44201400','3','0',NULL,'交警支队','','交警支队','2','冯伟健','','','','',NULL,'','1','1','2','6',NULL,NULL,'1','1','','','中山坦洲镇大兴酒店停车场出口','',113.474951,22.263601,'1','','',NULL,'0','2','',NULL,'',NULL,NULL,NULL,NULL,NULL,NULL,'2024-12-23',NULL,'1','1',0,0,0,0,0,0,0,0,0,0,0,0,0,30,NULL,NULL,'2','0',NULL,'A0201',0,'2025-12-18 02:12:42'),(178,30,'44201400071217950113',NULL,'中山坦洲镇大兴酒店停车场出口','44201400','3','0',NULL,'交警支队','1','交警支队','2','冯','','','','',NULL,'','1','1','2','6',NULL,NULL,'1','1','','','中山坦洲镇大兴酒店停车场出口','',113.474951,22.263601,'1','','',NULL,'0','2','',NULL,'',NULL,NULL,NULL,NULL,NULL,NULL,'2024-12-23',NULL,'1','1',0,0,0,0,0,0,0,0,0,0,0,0,0,30,NULL,NULL,'2','0',NULL,'A0201',0,'2025-12-18 02:12:42'),(185,32,'44200901001320004049',NULL,'4049六百六路19号龙发门业前-东往西-2','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442009010001','广东省中山市六百六路14号','广东省中山市六百六路14号',113.491241,22.623541,'1','1','',NULL,'0','2','44.103.75.85',NULL,'2c:a5:9c:46:4e:a2',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'A0699',2,'2025-12-18 02:28:01'),(186,32,'44200901001320004063',NULL,'4063浪源路2号接源村德恒学校路段-西南往东北-3','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442009010001','广东省中山市浪源路2号','中山市古镇时代慧童幼儿园',113.470867,22.603316,'1','3','',NULL,'0','2','44.103.75.115',NULL,'f8-a4-5f-ff-2d-4f',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'B0701',2,'2025-12-18 02:28:01'),(187,32,'44200903001320004031',NULL,'4031平一路16号','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442009010001','广东省中山市平一路16号','广东省中山市平一路16号',113.489111,22.680751,'1','2','',NULL,'0','2','44.103.75.69',NULL,'80:7c:62:b2:67:ae',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,0,0,0,0,0,0,0,30,NULL,NULL,'1','0',NULL,'A0699',2,'2025-12-18 02:28:01'),(188,32,'44200903001320004052',NULL,'4052义仓路1号三民学校东门-南往北-2','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,NULL,'1','2','442009010001','广东省中山市义仓路1号','广东省中山市义仓路1号',113.520631,22.620641,'1','1','',NULL,'0','2','44.103.75.88',NULL,'2c:a5:9c:3a:fe:46',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'B0701',2,'2025-12-18 02:28:01'),(189,32,'44200901001320004056',NULL,'4056护龙南路9号民江路中段-南往北1，2车道-2','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,NULL,'1','2','442009010001','广东省中山市护龙南路9号','广东省中山市护龙南路9号',113.479191,22.605581,'1','1','',NULL,'0','2','44.103.75.91',NULL,'2c:a5:9c:46:4e:8f',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0','省厅巡检离线，自查离线改停用','A0699',2,'2025-12-18 02:28:01'),(190,32,'44200901001320004066',NULL,'4066浪网大道145号浪网村委门前-西南往东北-3','442009','1','0',NULL,'广东省中山市民众分局','1','中山市公安局民众分局','2','梁鹏彬','139253162222','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','四期-01包-中国联通股份有限公司中山分公司-杨志全-18689390323/02包-高新兴-韩广雷-13302401405','1',NULL,'3','1','1,5','2','6','1',NULL,'1','1','2','442009010001','广东省中山市浪网大道145号','广东省中山市浪网大道145号',113.460561,22.629571,'1','3','',NULL,'0','2','44.103.75.101',NULL,'ec:c8:9c:6e:69:9a',NULL,NULL,'admin','hik-12345',NULL,'1','2022-07-30',NULL,'1','1',0,0,0,0,0,0,NULL,0,NULL,NULL,NULL,NULL,0,30,NULL,NULL,'1','0',NULL,'B0102',2,'2025-12-18 02:28:01'),(203,35,'44201001551310980269',NULL,'广东省中山市东凤镇金丰宾馆一楼门口（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇金丰宾馆一楼门口（治安）','广东省中山市东凤镇金丰宾馆一楼门口（治安）',113.262431,22.696601,'','','',NULL,NULL,'','44.105.13.149',NULL,'',NULL,NULL,'admin','Df@980071',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1001',0,'2025-12-18 13:05:32'),(204,35,'44201001551310980301',NULL,'广东省中山市东凤镇水南湾足浴三楼门口（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇水南湾足浴三楼门口（治安）','广东省中山市东凤镇水南湾足浴三楼门口（治安）',113.255789,22.723982,'','','',NULL,NULL,'','44.106.152.59',NULL,'',NULL,NULL,'admin','Df@980271',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1002',0,'2025-12-18 13:05:32'),(205,35,'44201001551310980303',NULL,'广东省中山市东凤镇沃尔客酒店一楼门口（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇沃尔客酒店一楼门口（治安）','广东省中山市东凤镇沃尔客酒店一楼门口（治安）',113.250557,22.726207,'','','',NULL,NULL,'','44.105.161.188',NULL,'',NULL,NULL,'admin','Df@980045',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1001',0,'2025-12-18 13:05:32'),(206,35,'44201001551310980262',NULL,'广东省中山市东凤镇濠鑫足浴二楼门口（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇濠鑫足浴二楼门口（治安）','广东省中山市东凤镇濠鑫足浴二楼门口（治安）',113.253895,22.686021,'','','',NULL,NULL,'','44.105.13.229',NULL,'',NULL,NULL,'admin','Df@980265',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1002',0,'2025-12-18 13:05:32'),(207,35,'44201001551310980270',NULL,'广东省中山市东凤镇六天快捷酒店一楼门口（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇六天快捷酒店一楼门口（治安）','广东省中山市东凤镇六天快捷酒店一楼门口（治安）',113.250241,22.703102,'','','',NULL,NULL,'','44.106.152.29',NULL,'',NULL,NULL,'admin','Df@980003',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1001',0,'2025-12-18 13:05:32'),(208,35,'44201001551310980265',NULL,'广东省中山市东凤镇万通酒店一楼大堂（治安）','442010','3',NULL,NULL,'东凤分局','','东凤分局','2','邓晓晴','15013049495','广东新逸达科技有限公司','达因天华-娱乐场所KTV人像-联系人-邓晓晴-联系电话-15013049495','',NULL,'',NULL,'2,5','','',NULL,NULL,NULL,NULL,'','','广东省中山市东凤镇万通酒店一楼大堂（治安）','广东省中山市东凤镇万通酒店一楼大堂（治安）',113.332977,22.736897,'','','',NULL,NULL,'','44.105.161.124',NULL,'',NULL,NULL,'admin','Df@980264',NULL,NULL,NULL,NULL,'1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,30,NULL,NULL,NULL,NULL,NULL,'B1001',0,'2025-12-18 13:05:32');
/*!40000 ALTER TABLE `audit_details` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-12-19 19:37:45
