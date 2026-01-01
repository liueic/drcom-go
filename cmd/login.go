package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagUser   string
	flagPass   string
	flagHost   string
	flagSave   bool
	flagNoSave bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录 Dr.COM 校园网",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("警告: 加载配置文件失败: %v\n", err)
			cfg = &config.Config{}
		}

		if flagHost != "" {
			cfg.Auth.Host = flagHost
		}
		if cfg.Auth.Host == "" {
			prompt := promptui.Prompt{
				Label:   "登录地址 (Host)",
				Default: "http://10.10.10.9:801",
			}
			res, err := prompt.Run()
			if err != nil {
				return
			}
			cfg.Auth.Host = res
		}
		viper.Set("auth.host", cfg.Auth.Host)

		if flagUser != "" {
			cfg.Auth.Username = flagUser
		}
		if cfg.Auth.Username == "" {
			prompt := promptui.Prompt{
				Label: "账号 (Username)",
			}
			res, err := prompt.Run()
			if err != nil {
				return
			}
			cfg.Auth.Username = res
		}
		viper.Set("auth.username", cfg.Auth.Username)

		if flagPass != "" {
			cfg.Auth.Password = flagPass
		}
		if cfg.Auth.Password == "" {
			prompt := promptui.Prompt{
				Label: "密码 (Password)",
				Mask:  '*',
			}
			res, err := prompt.Run()
			if err != nil {
				return
			}
			cfg.Auth.Password = res
		}
		viper.Set("auth.password", cfg.Auth.Password)

		// Decide whether to save config
		shouldSave := false
		if flagNoSave {
			shouldSave = false
		} else if flagSave {
			shouldSave = true
		} else if viper.ConfigFileUsed() != "" {
			// If config file exists/loaded, default to saving (updating) it
			shouldSave = true
		} else {
			// No flags and no existing config -> Prompt user
			prompt := promptui.Select{
				Label: "是否保存账号密码到本地? (公共电脑请选 No)",
				Items: []string{"Yes (保存 - 方便下次登录)", "No (不保存 - 公共电脑/临时使用)"},
			}
			_, result, err := prompt.Run()
			if err == nil && strings.HasPrefix(result, "Yes") {
				shouldSave = true
			}
		}

		if shouldSave {
			home, _ := os.UserHomeDir()
			configPath := home + "/.config/drcom-go/config.yaml"
			if viper.ConfigFileUsed() == "" {
				err := viper.WriteConfigAs(configPath)
				if err == nil {
					os.Chmod(configPath, 0600)
					fmt.Println("配置已保存至:", configPath)
				} else {
					fmt.Printf("保存配置失败: %v\n", err)
				}
			} else {
				viper.SafeWriteConfig()
			}
		}

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
		fmt.Println("正在登录...")
		resp, err := client.Login()
		if err != nil {
			fmt.Printf("登录请求失败: %v\n", err)
			return
		}

		if resp.Result == "1" || resp.Result == 1 || fmt.Sprintf("%v", resp.Result) == "1" {
			fmt.Printf("\033[32m登录接口成功: %s\033[0m\n", resp.Msg)
			verifyInternet()
		} else {
			// Check "Already online" case loosely
			if strings.Contains(resp.Msg, "已经在线") {
				fmt.Printf("\033[33m提示: %s\033[0m\n", resp.Msg)
				verifyInternet()
			} else {
				fmt.Printf("\033[31m登录失败: %s (返回码: %v)\033[0m\n", resp.Msg, resp.Result)
			}
		}
	},
}

func verifyInternet() {
	fmt.Print("正在验证外网连接...")
	time.Sleep(1 * time.Second)
	if drcom.CheckInternet() {
		fmt.Println("\033[32m [通过] (可以访问百度)\033[0m")
	} else {
		fmt.Println("\033[31m [失败] (无法访问百度，请检查网络设置或欠费状态)\033[0m")
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&flagUser, "user", "u", "", "校园网账号")
	loginCmd.Flags().StringVarP(&flagPass, "pass", "p", "", "校园网密码")
	loginCmd.Flags().StringVar(&flagHost, "host", "", "认证服务器地址 (例如 http://10.10.10.9:801)")
	loginCmd.Flags().BoolVar(&flagSave, "save", false, "强制保存配置到本地")
	loginCmd.Flags().BoolVar(&flagNoSave, "no-save", false, "不保存配置到本地")
}