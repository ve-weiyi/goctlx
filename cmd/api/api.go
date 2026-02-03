package api

import (
	"github.com/spf13/cobra"

	"github.com/ve-weiyi/goctlx/cmd/api/gin"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "生成 API 相关代码",
	}

	cmd.AddCommand(gin.NewGinCmd())
	return cmd
}
