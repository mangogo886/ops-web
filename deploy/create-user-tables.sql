-- 用户角色表
CREATE TABLE IF NOT EXISTS `user_role` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `role_name` varchar(50) NOT NULL COMMENT '角色名称',
  `role_code` tinyint(4) NOT NULL COMMENT '角色代码：0=管理员，1=普通用户',
  `permissions` text DEFAULT NULL COMMENT '角色权限（JSON格式，预留）',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_code` (`role_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色表';

-- 插入默认角色数据
INSERT INTO `user_role` (`role_name`, `role_code`, `permissions`) VALUES
('管理员', 0, '{"all": true}'),
('普通用户', 1, '{"view": true, "export": true}')
ON DUPLICATE KEY UPDATE `role_name`=VALUES(`role_name`);

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码（bcrypt加密）',
  `role_id` int(11) NOT NULL COMMENT '角色ID，关联user_role表',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_role_id` (`role_id`),
  CONSTRAINT `fk_user_role` FOREIGN KEY (`role_id`) REFERENCES `user_role` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

INSERT INTO `users` (`username`, `password`, `role_id`) VALUES
('admin', '$2a$10$iEyTdChlb/1.W73kXXHaluS3eHwjvLqqCNCNfUo2kp2myJLODHlKW', 1)
ON DUPLICATE KEY UPDATE `username`=VALUES(`username`);

-- 操作日志表
CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `username` varchar(50) NOT NULL COMMENT '操作用户',
  `action` varchar(255) NOT NULL COMMENT '操作描述',
  `ip` varchar(50) DEFAULT NULL COMMENT '来源IP',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户操作日志';

