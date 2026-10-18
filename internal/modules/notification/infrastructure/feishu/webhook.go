package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dujiao-next/internal/modules/notification/contract"
)

var hookIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}(-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}$`)

var webhookHTTPClient = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}

// SendWebhookMessage sends plain text to a Feishu custom group bot.
func (s *Sender) SendWebhookMessage(ctx context.Context, webhookURL, message string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	message = strings.TrimSpace(message)
	u, err := url.Parse(webhookURL)
	if err != nil || u.Scheme != "https" || u.Host != "open.feishu.cn" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || !strings.HasPrefix(u.Path, "/open-apis/bot/v2/hook/") || !hookIDPattern.MatchString(strings.TrimPrefix(u.Path, "/open-apis/bot/v2/hook/")) || message == "" {
		return contract.ErrConfigInvalid
	}
	body, err := json.Marshal(map[string]interface{}{"msg_type": "text", "content": map[string]string{"text": message}})
	if err != nil {
		return fmt.Errorf("%w: encode feishu webhook payload", contract.ErrSendFailed)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: create feishu webhook request", contract.ErrSendFailed)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := webhookHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: feishu webhook request failed", contract.ErrSendFailed)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: feishu webhook HTTP %d", contract.ErrSendFailed, resp.StatusCode)
	}
	var result struct {
		Code *int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Code == nil || *result.Code != 0 {
		return errors.New("feishu webhook rejected message")
	}
	return nil
}

var _ contract.FeishuWebhookSender = (*Sender)(nil)
