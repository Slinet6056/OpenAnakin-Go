package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL    = "https://api.anakin.ai"
	apiVersion = "2024-05-06"
)

// Message 定义聊天消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamCallback 定义流式回调接口
type StreamCallback interface {
	OnEvent(event, data string)
	OnComplete()
	OnError(err error)
}

// AnakinClient 定义Anakin客户端结构
type AnakinClient struct {
	httpClient  *http.Client
	modelAppIds map[string]int
}

// NewAnakinClient 创建新的Anakin客户端
func NewAnakinClient(modelAppIds map[string]int) *AnakinClient {
	return &AnakinClient{
		httpClient: &http.Client{
			Timeout: time.Second * 30,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConns:          100,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
			},
		},
		modelAppIds: modelAppIds,
	}
}

// SendMessage 发送消息并获取响应
func (c *AnakinClient) SendMessage(apiKey, model string, messages []Message) (string, error) {
	appID, err := c.getAppID(model)
	if err != nil {
		return "", err
	}

	content := c.buildMessageContent(messages)
	req, err := c.buildRequest(apiKey, appID, content, false)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

// SendStreamMessage 发送流式消息
func (c *AnakinClient) SendStreamMessage(apiKey, model string, messages []Message, callback StreamCallback) error {
	appID, err := c.getAppID(model)
	if err != nil {
		return err
	}

	content := c.buildMessageContent(messages)
	req, err := c.buildRequest(apiKey, appID, content, true)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		callback.OnError(fmt.Errorf("发送请求失败: %w", err))
		return err
	}

	go c.handleStreamResponse(resp, callback)
	return nil
}

// getAppID 获取应用ID
func (c *AnakinClient) getAppID(model string) (int, error) {
	appID, exists := c.modelAppIds[model]
	if !exists {
		return 0, fmt.Errorf("不支持的模型: %s", model)
	}
	return appID, nil
}

// buildMessageContent 构建消息内容
func (c *AnakinClient) buildMessageContent(messages []Message) string {
	var builder strings.Builder
	for _, msg := range messages {
		builder.WriteString(msg.Role)
		builder.WriteString(": ")
		builder.WriteString(msg.Content)
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

// buildRequest 构建请求
func (c *AnakinClient) buildRequest(apiKey string, appID int, content string, isStream bool) (*http.Request, error) {
	reqBody := map[string]interface{}{
		"content": content,
		"stream":  isStream,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chatbots/%d/messages", baseURL, appID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Anakin-Api-Version", apiVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	return req, nil
}

// handleResponse 处理非流式响应
func (c *AnakinClient) handleResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("请求失败: %s, 错误信息: %s", resp.Status, string(body))
	}

	var result struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Content, nil
}

// handleStreamResponse 处理流式响应
func (c *AnakinClient) handleStreamResponse(resp *http.Response, callback StreamCallback) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		callback.OnError(fmt.Errorf("请求失败: %s, 错误信息: %s", resp.Status, string(body)))
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				callback.OnError(fmt.Errorf("读取响应失败: %w", err))
			} else {
				callback.OnComplete()
			}
			return
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			callback.OnComplete()
			return
		}

		callback.OnEvent("message", data)
	}
}
