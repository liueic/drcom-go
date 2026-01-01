package drcom

import (
	"net/http"
	"time"
)

// CheckInternet attempts to connect to a reliable external website (Baidu)
// to verify actual internet connectivity.
func CheckInternet() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
    // We use a HEAD request to save bandwidth if possible, but GET is safer for some captive portals
    // that might intercept HEAD differently. GET is robust.
	resp, err := client.Get("https://www.baidu.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
