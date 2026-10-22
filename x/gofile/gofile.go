// Package gofile 把渲染好的 Go 源码整理后落盘。
package gofile

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"
)

// Write 创建目标目录，补齐并清理 import，随后按 0640 写入。
func Write(path string, src []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	formatted, err := imports.Process(path, src, nil)
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	if err := os.WriteFile(path, formatted, 0640); err != nil {
		return err
	}

	fmt.Println("✅ generated:", path)
	return nil
}
