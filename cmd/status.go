package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看网络状态 (漂亮面板版)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("无法加载配置: %v\n", err)
			return
		}

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
		res, err := client.GetStatus()
		if err != nil {
			color.Red("❌ 获取状态失败: %v", err)
			return
		}

		var flowMB, money float64
		var userName string = cfg.Auth.Username

		if len(res.Data) > 0 {
			flowMB = res.Data[0].UserFlow
			money = res.Data[0].UserMoney
		} else if res.UserInfo.UserFlow != "" {
			flowKB, _ := strconv.ParseFloat(res.UserInfo.UserFlow, 64)
			flowMB = flowKB / 1024
			money, _ = strconv.ParseFloat(res.UserInfo.UserBalance, 64)
			userName = res.UserInfo.UserName
		} else {
			color.Yellow("⚠️ 未获取到有效状态信息。请检查登录状态。\n")
			return
		}

		flowGB := flowMB / 1024
		threshold := cfg.Alert.TrafficThreshold
		if threshold == 0 {
			threshold = 80.0
		}

		fmt.Println("\n" + color.CyanString("📡 Dr.COM 状态面板"))
		fmt.Println(strings.Repeat("-", 35))

		fmt.Printf("👤 账号: %s\n", userName)
		fmt.Printf("💰 余额: %.2f 元\n", money)
		
		trafficStr := fmt.Sprintf("%.2f GB", flowGB)
		if flowGB > threshold*0.9 {
			trafficStr = color.RedString(trafficStr + " [危险]")
		} else if flowGB > threshold*0.7 {
			trafficStr = color.YellowString(trafficStr + " [注意]")
		} else {
			trafficStr = color.GreenString(trafficStr)
		}
		fmt.Printf("📊 流量: %s\n", trafficStr)

		// Progress bar
		printProgressBar(flowGB, threshold)

		if flowGB >= threshold {
			color.Red("\n⚠️  警告: 流量已达上限 (阈值: %.2f GB)", threshold)
		} else if flowGB > threshold*0.8 {
			color.Yellow("\n⚠️  提示: 流量接近上限 (阈值: %.2f GB)", threshold)
		}
		fmt.Println(strings.Repeat("-", 35))
	},
}

func printProgressBar(current, total float64) {
	width := 25
	percent := current / total
	if percent > 1 {
		percent = 1
	}
	filled := int(float64(width) * percent)
	if filled < 0 { filled = 0 }
	bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
	
	fmt.Printf("[%s] %.0f%%\n", bar, (current/total)*100)
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
