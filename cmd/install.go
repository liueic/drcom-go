package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "一键安装为 Systemd 服务 (仅支持 Linux)",
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			color.Red("❌ 该功能仅支持 Linux 系统。")
			return
		}

		if os.Geteuid() != 0 {
			color.Yellow("⚠️ 需要 root 权限。请尝试: sudo ./drcom install")
			return
		}

		exePath, err := os.Executable()
		if err != nil {
			color.Red("❌ 无法获取程序路径: %v", err)
			return
		}
		exePath, _ = filepath.Abs(exePath)

		serviceContent := fmt.Sprintf(`[Unit]
Description=Dr.COM Daemon Service
After=network-online.target syslog.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s daemon
Restart=always
RestartSec=10
User=root

[Install]
WantedBy=multi-user.target
`, exePath)

		servicePath := "/etc/systemd/system/drcom.service"
		err = os.WriteFile(servicePath, []byte(serviceContent), 0644)
		if err != nil {
			color.Red("❌ 写入服务文件失败: %v", err)
			return
		}

		fmt.Println("✅ 已创建服务文件:", servicePath)

		// Reload systemd
		exec.Command("systemctl", "daemon-reload").Run()
		// Enable service
		exec.Command("systemctl", "enable", "drcom").Run()
		// Start service
		err = exec.Command("systemctl", "start", "drcom").Run()
		if err != nil {
			color.Red("❌ 启动服务失败: %v", err)
			return
		}

		color.Green("🚀 服务已成功安装并启动！")
		fmt.Println("使用 'systemctl status drcom' 查看状态")
		fmt.Println("使用 'journalctl -u drcom -f' 查看日志")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
