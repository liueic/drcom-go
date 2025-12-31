package drcom

// API Response Wrapper
type ApiResponse struct {
	Code int         `json:"code"` // 200 success, 500 error
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Status Data Structure for API
type ApiStatusData struct {
	Success  bool    `json:"success"`
	Username string  `json:"username"`
	FlowGB   float64 `json:"flow_gb"`
	Fee      float64 `json:"fee"`
	IP       string  `json:"ip"`
    Message  string  `json:"message,omitempty"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
