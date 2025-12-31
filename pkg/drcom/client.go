package drcom

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func NewClient(host, username, password string) *DrComClient {
	return &DrComClient{
		Host:     strings.TrimRight(host, "/"),
		Username: username,
		Password: password,
	}
}

// Auto-detect local IP if not set
func (c *DrComClient) GetLocalIP() string {
	if c.IP != "" {
		return c.IP
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	c.IP = localAddr.IP.String()
	return c.IP
}

func (c *DrComClient) Login() (*LoginResponse, error) {
	api := c.Host + "/eportal/portal/login"
	
	// Prepare params
	params := url.Values{}
	callback := fmt.Sprintf("dr%d", 1000+rand.Intn(9000))
	params.Set("callback", callback)
	params.Set("login_method", "1")
	// Based on request.md, there is a weird prefix: ,`, (comma backtick comma)
	// We will try to prepend it if the username doesn't already have it.
	// Actually, let's just prepend it as per the capture. 
	// If this varies, we might need a config option.
	// Decoding %2C%60%2C -> ,`,
	params.Set("user_account", ",`,"+c.Username) 
	params.Set("user_password", c.Password)
	params.Set("wlan_user_ip", c.GetLocalIP())
	params.Set("wlan_user_mac", "000000000000")
	params.Set("jsVersion", "4.2.1")
	params.Set("terminal_type", "1")
	params.Set("lang", "zh-cn")
	params.Set("v", strconv.Itoa(rand.Intn(9999)))

	reqURL := api + "?" + params.Encode()
	
	resp, err := c.doRequest(reqURL)
	if err != nil {
		return nil, err
	}

	var res LoginResponse
	if err := parseJSONP(resp, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DrComClient) Logout() error {
	api := c.Host + "/eportal/portal/logout"
	params := url.Values{}
	callback := fmt.Sprintf("dr%d", 1000+rand.Intn(9000))
	params.Set("callback", callback)
	params.Set("user_account", c.Username) // Logout usually doesn't need the prefix
	params.Set("wlan_user_ip", c.GetLocalIP())
	params.Set("jsVersion", "4.2.1")
	params.Set("v", strconv.Itoa(rand.Intn(9999)))
    
    // Some versions use login_method=1 for logout too? No, usually distinct endpoint.
    
    reqURL := api + "?" + params.Encode()
	resp, err := c.doRequest(reqURL)
	if err != nil {
		return err
	}
    // We just check if it returns valid JSONP, ignoring content usually
    var res map[string]interface{}
    return parseJSONP(resp, &res)
}

func (c *DrComClient) GetStatus() (*UserInfoResponse, error) {
	// Using loadUserInfo as it seems richer
	api := c.Host + "/eportal/portal/custom/loadUserInfo"
	params := url.Values{}
	callback := fmt.Sprintf("dr%d", 1000+rand.Intn(9000))
	params.Set("callback", callback)
	params.Set("wlan_user_ip", c.GetLocalIP())
	params.Set("is_login", "0") // request.md has is_login=0, maybe checks if logged in?
	params.Set("jsVersion", "4.2.1")
	params.Set("v", strconv.Itoa(rand.Intn(9999)))
	params.Set("lang", "zh")

	reqURL := api + "?" + params.Encode()
	resp, err := c.doRequest(reqURL)
	if err != nil {
		return nil, err
	}

	var res UserInfoResponse
	if err := parseJSONP(resp, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DrComClient) doRequest(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	
	// Headers from request.md
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Referer", c.Host + "/")
    
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func parseJSONP(content string, v interface{}) error {
	// Match content inside callback(...)
	re := regexp.MustCompile(`^.*?\((.*)\);?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(content))
	
	jsonStr := content
	if len(matches) > 1 {
		jsonStr = matches[1]
	}
    
    // Debug print
    // fmt.Println("DEBUG JSON:", jsonStr)
    
    // Sometimes JSONP is just callback({ ... }) without ;
    // Or sometimes failure is just html.
    
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		// Log content for debugging if it fails
		return fmt.Errorf("failed to parse JSON: %v, content: %s", err, content)
	}
	return nil
}
