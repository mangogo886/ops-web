package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
)

// 用于生成管理员密码的 bcrypt hash
// 使用方法: go run scripts/init-admin.go <password>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run scripts/init-admin.go <password>")
		fmt.Println("示例: go run scripts/init-admin.go admin123")
		os.Exit(1)
	}

	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("生成密码hash失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("密码: %s\n", password)
	fmt.Printf("BCrypt Hash: %s\n", string(hash))
	fmt.Printf("\nSQL插入/更新语句（管理员用户）:\n")
	fmt.Printf("-- 如果用户不存在，插入新用户\n")
	fmt.Printf("INSERT INTO `users` (`username`, `password`, `role_id`) VALUES\n")
	fmt.Printf("('admin', '%s', 1)\n", string(hash))
	fmt.Printf("ON DUPLICATE KEY UPDATE `password`=VALUES(`password`);\n")
	fmt.Printf("\n-- 或者直接更新现有用户密码\n")
	fmt.Printf("UPDATE `users` SET `password`='%s' WHERE `username`='admin';\n", string(hash))
	fmt.Printf("\n注意：role_id=1 对应 user_role 表中 id=1 的记录（role_code=0的管理员角色）\n")
}























