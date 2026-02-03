package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cast"
)

func main() {
	// 定义固定前缀和后缀
	prefix := "1880771" // 南宁青秀区核心号段前缀
	suffix := "26"      // 固定尾号
	var phoneList []string

	// 生成00-99的2位中间数，拼接完整手机号
	for i := 0; i < 100; i++ {
		// %02d 确保数字补零（如1变成01，9变成09）
		mid := fmt.Sprintf("%02d", i)
		fullPhone := prefix + mid + suffix
		phoneList = append(phoneList, cast.ToString(i)+": "+fullPhone)
	}

	// 将号码写入文件（每行一个）
	content := strings.Join(phoneList, "\n")
	err := os.WriteFile("qingxiu_188_phones.txt", []byte(content), 0644)
	if err != nil {
		fmt.Printf("文件写入失败：%v\n", err)
		return
	}

	// 打印执行结果
	fmt.Printf("✅ 成功生成 %d 个手机号（1880771xx26）\n", len(phoneList))
	fmt.Printf("📄 号码已保存到当前目录的 qingxiu_188_phones.txt 文件\n")
	fmt.Println("\n📌 前5个号码示例：")
	for i := 0; i < 5; i++ {
		fmt.Printf("   %s\n", phoneList[i])
	}
}
