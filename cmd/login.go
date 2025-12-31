package cmd

import (
	"fmt"
	"os"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagUser string
	flagPass string
	flagHost string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录 Dr.COM 校园网",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			// If config fails (e.g. permissions), we just warn and proceed with defaults/flags
			fmt.Printf("警告: 加载配置文件失败: %v\n", err)
			cfg = &config.Config{} 
		}

		// Priority: Flags > Config > Interactive Prompt

		// Host
		if flagHost != "" {
			cfg.Auth.Host = flagHost
		}
		if cfg.Auth.Host == "" {
			prompt := promptui.Prompt{
				Label:   "登录地址 (Host)",
				Default: "http://10.10.10.9:801", // Updated based on curl port 801
			}
			res, err := prompt.Run()
			if err != nil {
				return
			}
			cfg.Auth.Host = res
		}
        viper.Set("auth.host", cfg.Auth.Host)

		// Username
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

		// Password
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

        // Auto save config
        home, _ := os.UserHomeDir()
        configPath := home + "/.config/drcom-go/config.yaml"
        if viper.ConfigFileUsed() == "" {
             viper.WriteConfigAs(configPath)
             os.Chmod(configPath, 0600)
             fmt.Println("配置已保存至:", configPath)
        } else {
             viper.SafeWriteConfig()
        }

		client := drcom.NewClient(cfg.Auth.Host, cfg.Auth.Username, cfg.Auth.Password)
		fmt.Println("正在登录...")
		resp, err := client.Login()
		if err != nil {
			fmt.Printf("登录请求失败: %v\n", err)
			return
		}

        // Check success
		if resp.Result == "1" || resp.Result == 1 || fmt.Sprintf("%v", resp.Result) == "1" {
			fmt.Printf("\033[32m登录成功: %s\033[0m\n", resp.Msg)
		} else {
			fmt.Printf("\033[31m登录失败: %s (返回码: %v)\033[0m\n", resp.Msg, resp.Result)
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&flagUser, "user", "u", "", "校园网账号")
	loginCmd.Flags().StringVarP(&flagPass, "pass", "p", "", "校园网密码")
	loginCmd.Flags().StringVar(&flagHost, "host", "", "认证服务器地址 (例如 http://10.10.10.9:801)")
}