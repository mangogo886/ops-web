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
-- Table structure for table `audit_tasks`
--

DROP TABLE IF EXISTS `audit_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `audit_tasks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `file_name` varchar(255) NOT NULL COMMENT '档案名称',
  `organization` varchar(100) NOT NULL COMMENT '机构/子公司名称',
  `import_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '导入时间',
  `audit_status` varchar(20) NOT NULL DEFAULT '待审核' COMMENT '审核状态：待审核、已审核待整改、已完成',
  `record_count` int(11) NOT NULL DEFAULT '0' COMMENT '导入记录数量',
  `is_single_soldier` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否单兵设备：0-否，1-是',
  `audit_comment` text COMMENT '审核意见',
  `auditor` varchar(50) DEFAULT NULL COMMENT '审核人',
  `audit_time` timestamp NULL DEFAULT NULL COMMENT '审核时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `archive_type` varchar(50) DEFAULT NULL COMMENT '档案类型：新增、取推、补档案',
  PRIMARY KEY (`id`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_organization` (`organization`),
  KEY `idx_import_time` (`import_time`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COMMENT='档案审核任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `audit_tasks`
--

LOCK TABLES `audit_tasks` WRITE;
/*!40000 ALTER TABLE `audit_tasks` DISABLE KEYS */;
INSERT INTO `audit_tasks` (`id`, `file_name`, `organization`, `import_time`, `audit_status`, `record_count`, `is_single_soldier`, `audit_comment`, `auditor`, `audit_time`, `created_at`, `updated_at`, `archive_type`) VALUES (29,'横栏分局卡口设备测试数据','横栏分局','2025-12-18 02:12:21','未审核',4,0,NULL,NULL,NULL,'2025-12-18 02:12:21','2025-12-18 02:12:21','新增'),(30,'交警支队卡口设备测试数据','交警支队','2025-12-18 02:12:43','未审核',2,0,NULL,NULL,NULL,'2025-12-18 02:12:42','2025-12-18 02:12:42','新增'),(32,'民众分局卡口设备测试数据 - 副本','民众分局','2025-12-18 02:17:00','已完成',6,0,NULL,NULL,NULL,'2025-12-18 02:17:00','2025-12-18 02:28:01','新增'),(35,'东凤分局设备测试数据','东凤分局','2025-12-18 13:05:33','未审核',6,0,'已处理',NULL,NULL,'2025-12-18 13:05:32','2025-12-19 05:02:39','取推');
/*!40000 ALTER TABLE `audit_tasks` ENABLE KEYS */;
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
