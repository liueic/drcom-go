package drcom

// Old structs (kept if needed, but defining new ones for current observation)

type LoginResponse struct {
	Result   interface{} `json:"result"`   // "1" or 1 usually means success
	Msg      string      `json:"msg"`      // "登录成功"
	RetCode  interface{} `json:"ret_code"` // 2 usually
}

// Updated based on actual response
type UserInfoResponse struct {
    Code string `json:"code"` // "1"
    Msg string `json:"msg"`
    Data []UserData `json:"data"`
    // Keep old fields just in case it varies
	Result    interface{} `json:"result"`
	UserInfo  UserInfo    `json:"user_info"`
}

type UserData struct {
    UserFlow float64 `json:"USERFLOW"` // It was number in JSON
    UserMoney float64 `json:"USERMONEY"`
    UserTime int `json:"USERTIME"`
}

type UserInfo struct {
	UserIndex   string `json:"userIndex"`
	UserAccount string `json:"userAccount"`
	UserName    string `json:"userName"`
	UserBalance string `json:"userBalance"`
	UserFlow    string `json:"userFlow"`
}

type DrComClient struct {
	Host     string
	Username string
	Password string
	IP       string // Local IP
}