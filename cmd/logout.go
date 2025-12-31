package cmd

import (
	"fmt"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "注销/退出校园网",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("无法加载配置: %v\n", err)
			return
		}

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
		err = client.Logout()
		if err != nil {
			fmt.Printf("注销失败: %v\n", err)
			return
		}
        fmt.Println("注销请求已发送。")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}