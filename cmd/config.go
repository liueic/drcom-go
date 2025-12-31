package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "设置配置项",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			color.Yellow("用法: drcom config set <key> <value>")
			return
		}
		key := args[0]
		val := args[1]
		viper.Set(key, val)
		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}
		if err != nil {
			color.Red("❌ 保存配置失败: %v", err)
			return
		}
		color.Green("✅ 已设置 %s = %s", key, val)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		fmt.Println(color.CyanString("当前配置项:"))
		for k, v := range settings {
			fmt.Printf("  %s: %v\n", k, v)
		}
		fmt.Printf("\n配置文件路径: %s\n", viper.ConfigFileUsed())
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}
