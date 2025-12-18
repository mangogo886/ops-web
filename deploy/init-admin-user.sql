-- ============================================
-- 初始化管理员账户
-- 说明: 此脚本用于创建初始管理员账户
-- ============================================

-- 1. 插入角色数据（如果不存在）
INSERT IGNORE INTO `user_role` (`id`, `role_name`, `role_code`) VALUES
(1, '管理员', 0),
(2, '普通用户', 1);

-- 2. 创建初始管理员账户
-- 默认用户名: admin
-- 默认密码: admin123
-- 密码已使用bcrypt加密（cost=10）
-- 如果账户已存在，此语句不会报错（使用INSERT IGNORE）

-- 注意：以下密码hash是 "admin123" 的bcrypt加密结果
-- 如果需要修改默认密码，可以使用以下Go代码生成新的hash：
-- package main
-- import (
--     "fmt"
--     "golang.org/x/crypto/bcrypt"
-- )
-- func main() {
--     hash, _ := bcrypt.GenerateFromPassword([]byte("your_password"), bcrypt.DefaultCost)
--     fmt.Println(string(hash))
-- }

INSERT IGNORE INTO `users` (`id`, `username`, `password`, `role_id`) VALUES
(1, 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1);

-- 说明：
-- 1. 默认管理员账户：用户名 admin，密码 admin123
-- 2. 首次登录后请立即修改密码！
-- 3. 如需创建其他用户，请登录系统后在"用户信息"页面添加


