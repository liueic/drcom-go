package cmd

import (
	"fmt"
	"net/http"
	"time"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "后台守护进程 (带 Webhook 通知)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("无法加载配置: %v\n", err)
			return
		}

		if cfg.Auth.Username == "" || cfg.Auth.Password == "" {
			fmt.Println("请先登录配置账号信息ảng")
			return
		}

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
        interval := time.Duration(cfg.Daemon.Interval) * time.Second
        if interval == 0 {
            interval = 60 * time.Second
        }

		color.Cyan("🚀 守护进程已启动 (检测间隔: %v)...", interval)
        
        lastAlertTime := time.Time{}

		for {
            if !checkInternet() {
                color.Yellow("[守护进程] 网络断开。正在尝试重连...")
                resp, err := client.Login()
                if err != nil {
                    color.Red("[错误] 登录请求失败: %v", err)
                } else {
                    success := resp.Result == "1" || resp.Result == 1 || fmt.Sprintf("%v", resp.Result) == "1"
                    // Also count "Already online" as success or at least handled
                    if success || (resp.Msg != "" && (resp.Msg == "已经在线" || fmt.Sprintf("%v", resp.Msg) == "已经在线")) {
                         color.Green("[成功] 重新连接成功: %s", resp.Msg)
                         drcom.SendWebhook(cfg.Alert.WebhookURL, "网络已重连: "+resp.Msg)
                    } else {
                         color.Red("[失败] 登录失败: %s", resp.Msg)
                    }
                }
            }
            
            // Traffic Check (once per hour to avoid spam)
            if time.Since(lastAlertTime) > 1*time.Hour {
                res, err := client.GetStatus()
                if err == nil {
                    var flowMB float64
                    if len(res.Data) > 0 {
                        flowMB = res.Data[0].UserFlow
                    }
                    flowGB := flowMB / 1024
                    threshold := cfg.Alert.TrafficThreshold
                    if threshold > 0 && flowGB >= threshold {
                        drcom.SendWebhook(cfg.Alert.WebhookURL, fmt.Sprintf("⚠️ 流量警告: 当前已用 %.2f GB, 超过阈值 %.2f GB", flowGB, threshold))
                        lastAlertTime = time.Now()
                    }
                }
            }

			time.Sleep(interval)
		}
	},
}

func checkInternet() bool {
    client := http.Client{
        Timeout: 3 * time.Second,
    }
    _, err := client.Get("http://www.baidu.com")
    return err == nil
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
