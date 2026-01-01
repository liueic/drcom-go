package cmd

import (
	"fmt"
	"strings"
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
			fmt.Println("请先登录配置账号信息。")
			return
		}

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
        interval := time.Duration(cfg.Daemon.Interval) * time.Second
        if interval == 0 {
            interval = 60 * time.Second
        }

		color.Cyan("🚀 守护进程已启动 (检测间隔: %v)...", interval)
        
        lastAlertTime := time.Time{}
        lastStatusLogTime := time.Time{}

		for {
            isOnline := drcom.CheckInternet()
            
            if !isOnline {
                color.Yellow("[%s] 网络断开。正在尝试重连...", time.Now().Format("15:04:05"))
                resp, err := client.Login()
                if err != nil {
                    color.Red("[错误] 登录请求失败: %v", err)
                } else {
                    // Check strict success
                    success := resp.Result == "1" || resp.Result == 1 || fmt.Sprintf("%v", resp.Result) == "1"
                    alreadyOnline := (resp.Msg != "" && strings.Contains(resp.Msg, "已经在线"))
                    
                    if success || alreadyOnline {
                         // Double check internet
                         time.Sleep(1 * time.Second) // Wait a sec for NAT/Rule propagation
                         if drcom.CheckInternet() {
                             color.Green("[成功] 重新连接成功: %s (且外网可达)", resp.Msg)
                             drcom.SendWebhook(cfg.Alert.WebhookURL, "网络已重连: "+resp.Msg)
                         } else {
                             color.Red("[警告] 登录接口返回成功，但外网依然不可达！")
                         }
                    } else {
                         color.Red("[失败] 登录失败: %s", resp.Msg)
                    }
                }
            }
            
            // Periodic Status Update (Log every 10 mins or so, Alert on Threshold)
            // We verify status even if online to update logs/monitor flow
            if time.Since(lastStatusLogTime) > 10*time.Minute || (!isOnline) {
                res, err := client.GetStatus()
                if err == nil {
                    var flowMB float64
                    if len(res.Data) > 0 {
                        flowMB = res.Data[0].UserFlow
                    } else if res.UserInfo.UserFlow != "" {
                         // parsing logic fallback...
                    }
                    
                    flowGB := flowMB / 1024
                    if isOnline {
                        fmt.Printf("[%s] 状态正常 | 流量: %.2f GB | 余额: %.2f\n", 
                            time.Now().Format("15:04"), flowGB, res.Data[0].UserMoney)
                    }
                    lastStatusLogTime = time.Now()

                    // Threshold Alert (Keep hourly restriction to avoid spam)
                    threshold := cfg.Alert.TrafficThreshold
                    if threshold > 0 && flowGB >= threshold && time.Since(lastAlertTime) > 1*time.Hour {
                        msg := fmt.Sprintf("⚠️ 流量警告: 当前已用 %.2f GB, 超过阈值 %.2f GB", flowGB, threshold)
                        color.Red(msg)
                        drcom.SendWebhook(cfg.Alert.WebhookURL, msg)
                        lastAlertTime = time.Now()
                    }
                }
            }

			time.Sleep(interval)
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
