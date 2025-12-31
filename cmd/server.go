package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"drcom-go/pkg/config"
	"drcom-go/pkg/drcom"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serverPort string
	configLock sync.Mutex
    globalCfg  *config.Config
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 HTTP API 服务 (Microservice)",
	Run:   runServer,
}

func init() {
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "", "服务监听端口 (默认使用配置)")
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) {
	var err error
	globalCfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("无法加载配置: %v\n", err)
		return
	}

    // Override port if flag is set
	if serverPort != "" {
		globalCfg.Server.Port = serverPort
	}
    if globalCfg.Server.Port == "" {
        globalCfg.Server.Port = "8080"
    }

	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/logout", handleLogout)
    
    // Simple Dashboard
    http.HandleFunc("/", handleDashboard)

	port := globalCfg.Server.Port
	color.Green("🌐 Dr.COM API Server listening on :%s", port)
    color.Cyan("   ➜ Dashboard: http://localhost:%s/", port)
    color.Cyan("   ➜ API:       http://localhost:%s/api/status", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		color.Red("启动失败: %v", err)
	}
}

func getClient() *drcom.DrComClient {
    configLock.Lock()
    defer configLock.Unlock()
    return drcom.NewClient(globalCfg.Auth.Host, globalCfg.Auth.Username, globalCfg.Auth.Password)
}

func checkToken(r *http.Request) bool {
    // If token is configured, check it
    if globalCfg.Server.Token != "" {
        token := r.Header.Get("X-API-Token")
        if token == "" {
             token = r.URL.Query().Get("token")
        }
        return token == globalCfg.Server.Token
    }
    return true
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
    if !checkToken(r) {
        http.Error(w, "Forbidden", 403)
        return
    }

	client := getClient()
	res, err := client.GetStatus()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")

	if err != nil {
		json.NewEncoder(w).Encode(drcom.ApiResponse{Code: 500, Msg: err.Error()})
		return
	}

	data := drcom.ApiStatusData{
        Success: true, // If we got status, we are likely fine
        IP: client.GetLocalIP(),
    }
    
    // Parse Logic reused from status command
    if len(res.Data) > 0 {
        data.FlowGB = res.Data[0].UserFlow / 1024
        data.Fee = res.Data[0].UserMoney
        data.Username = globalCfg.Auth.Username // or res.UserAccount if available
    } else if res.UserInfo.UserFlow != "" {
        flowKB, _ := strconv.ParseFloat(res.UserInfo.UserFlow, 64)
        data.FlowGB = flowKB / (1024 * 1024)
        data.Fee, _ = strconv.ParseFloat(res.UserInfo.UserBalance, 64)
        data.Username = res.UserInfo.UserName
    } else {
        data.Success = false
        data.Message = "Empty data received"
    }

	json.NewEncoder(w).Encode(drcom.ApiResponse{Code: 200, Msg: "success", Data: data})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
    if !checkToken(r) {
        http.Error(w, "Forbidden", 403)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", 405)
        return
    }

    var req drcom.LoginRequest
    if r.Header.Get("Content-Type") == "application/json" {
         json.NewDecoder(r.Body).Decode(&req)
    }

    configLock.Lock()
    if req.Username != "" && req.Password != "" {
        globalCfg.Auth.Username = req.Username
        globalCfg.Auth.Password = req.Password
        // Save to runtime viper so it persists? Or just runtime?
        // Let's update viper too
        viper.Set("auth.username", req.Username)
        viper.Set("auth.password", req.Password)
        // Optionally save to disk: config.SaveConfig(globalCfg)
    }
    configLock.Unlock()

    client := getClient()
    resp, err := client.Login()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    apiResp := drcom.ApiResponse{Code: 200, Msg: "Login executed"}
    
    if err != nil {
        apiResp.Code = 500
        apiResp.Msg = err.Error()
    } else {
        apiResp.Data = resp
        if resp.Result == "1" || resp.Result == 1 || fmt.Sprintf("%v", resp.Result) == "1" {
             apiResp.Msg = "Login Success: " + resp.Msg
        } else {
             // Handle "Already online" as non-error technically?
             apiResp.Msg = "Login Result: " + resp.Msg
        }
    }
    json.NewEncoder(w).Encode(apiResp)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
    if !checkToken(r) {
        http.Error(w, "Forbidden", 403)
        return
    }
    
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", 405)
        return
    }

    client := getClient()
    err := client.Logout()
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if err != nil {
        json.NewEncoder(w).Encode(drcom.ApiResponse{Code: 500, Msg: err.Error()})
    } else {
        json.NewEncoder(w).Encode(drcom.ApiResponse{Code: 200, Msg: "Logout signal sent"})
    }
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
    html := `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dr.COM 控制台</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background: #f0f2f5; display: flex; justify-content: center; padding-top: 50px; }
        .card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); width: 400px; text-align: center; }
        h1 { margin-bottom: 20px; color: #333; }
        .stat { font-size: 18px; margin: 10px 0; display: flex; justify-content: space-between; }
        .stat-value { font-weight: bold; color: #007bff; }
        .btn { padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 16px; margin: 10px; color: white; transition: background 0.2s; }
        .btn-login { background: #28a745; }
        .btn-login:hover { background: #218838; }
        .btn-logout { background: #dc3545; }
        .btn-logout:hover { background: #c82333; }
        .refresh { font-size: 14px; color: #666; cursor: pointer; text-decoration: underline; margin-top: 10px; display: inline-block; }
    </style>
</head>
<body>
    <div class="card">
        <h1>📡 Dr.COM 面板</h1>
        <div id="loading">加载中...</div>
        <div id="content" style="display:none;">
            <div class="stat"><span>👤 账号:</span> <span class="stat-value" id="user">-</span></div>
            <div class="stat"><span>💰 余额:</span> <span class="stat-value" id="fee">-</span></div>
            <div class="stat"><span>📊 流量:</span> <span class="stat-value" id="flow">-</span></div>
            <hr style="border:0; border-top:1px solid #eee; margin: 20px 0;">
            <button class="btn btn-login" onclick="doAction('login')">重新登录</button>
            <button class="btn btn-logout" onclick="doAction('logout')">注销</button>
        </div>
        <div class="refresh" onclick="fetchStatus()">刷新状态</div>
    </div>

    <script>
        const API_BASE = "/api"; 
        // If you set a token in config, append ?token=YOUR_TOKEN to the URL when visiting
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token') || '';

        function getHeaders() {
            return token ? { 'X-API-Token': token } : {};
        }

        async function fetchStatus() {
            document.getElementById('loading').style.display = 'block';
            document.getElementById('content').style.display = 'none';
            try {
                const res = await fetch(API_BASE + '/status?token=' + token, { headers: getHeaders() });
                const json = await res.json();
                if (json.code === 200 && json.data) {
                    document.getElementById('user').innerText = json.data.username || '未知';
                    document.getElementById('fee').innerText = json.data.fee.toFixed(2) + ' 元';
                    document.getElementById('flow').innerText = json.data.flow_gb.toFixed(2) + ' GB';
                } else {
                    alert('获取状态失败: ' + json.msg);
                }
            } catch (e) {
                console.error(e);
                document.getElementById('user').innerText = '离线';
            } finally {
                document.getElementById('loading').style.display = 'none';
                document.getElementById('content').style.display = 'block';
            }
        }

        async function doAction(action) {
            if (!confirm('确定要执行 ' + action + ' 吗?')) return;
            try {
                const res = await fetch(API_BASE + '/' + action + '?token=' + token, { 
                    method: 'POST',
                    headers: getHeaders()
                });
                const json = await res.json();
                alert(json.msg);
                fetchStatus();
            } catch (e) {
                alert('操作失败');
            }
        }

        fetchStatus();
    </script>
</body>
</html>`
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(html))
}
