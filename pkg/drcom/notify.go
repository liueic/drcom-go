package drcom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookPayload struct {
	MsgType string `json:"msg_type"` // For Feishu
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
    // For DingTalk style
    Text struct {
        Content string `json:"content"`
    } `json:"text"`
}

func SendWebhook(url, message string) error {
	if url == "" {
		return nil
	}

	// Simple generic payload that works with most (using Feishu/DingTalk style)
    // We'll just try a basic text payload
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": "[Dr.COM] " + message,
		},
        "content": map[string]string{ // Feishu compat
            "text": "[Dr.COM] " + message,
        },
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
