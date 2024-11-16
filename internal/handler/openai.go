package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Slinet6056/OpenAnakin-Go/internal/client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OpenAIRequest OpenAI请求结构
type OpenAIRequest struct {
	Model    string           `json:"model"`
	Messages []client.Message `json:"messages"`
	Stream   bool             `json:"stream"`
}

// OpenAIResponse OpenAI响应结构
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Delta   `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Delta struct {
	Content string `json:"content"`
}

type OpenAIHandler struct {
	anakinClient *client.AnakinClient
}

func NewOpenAIHandler(anakinClient *client.AnakinClient) *OpenAIHandler {
	return &OpenAIHandler{
		anakinClient: anakinClient,
	}
}

// ChatCompletions 处理聊天完成请求
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	var req OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息列表不能为空"})
		return
	}

	apiKey := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	if req.Stream {
		h.handleStreamRequest(c, apiKey, &req)
	} else {
		h.handleNonStreamRequest(c, apiKey, &req)
	}
}

// handleNonStreamRequest 处理非流式请求
func (h *OpenAIHandler) handleNonStreamRequest(c *gin.Context, apiKey string, req *OpenAIRequest) {
	response, err := h.anakinClient.SendMessage(apiKey, req.Model, req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("发送消息失败: %v", err)})
		return
	}

	openAIResp := h.convertToOpenAIResponse(response, req.Model)
	c.JSON(http.StatusOK, openAIResp)
}

// handleStreamRequest 处理流式请求
func (h *OpenAIHandler) handleStreamRequest(c *gin.Context, apiKey string, req *OpenAIRequest) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	var wg sync.WaitGroup
	wg.Add(1)

	err := h.anakinClient.SendStreamMessage(apiKey, req.Model, req.Messages, &streamCallback{
		context: c,
		model:   req.Model,
		done:    &wg,
	})

	if err != nil {
		wg.Done()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("处理流式请求失败: %v", err)})
		return
	}

	wg.Wait()
}

// streamCallback 实现 StreamCallback 接口
type streamCallback struct {
	context *gin.Context
	model   string
	done    *sync.WaitGroup
}

func (cb *streamCallback) OnEvent(event, data string) {
	openAIFormatData := convertToOpenAIFormat(data, cb.model)
	if openAIFormatData != "" {
		cb.context.Writer.Write([]byte("data: " + openAIFormatData + "\n\n"))
		cb.context.Writer.Flush()
	}
}

func (cb *streamCallback) OnComplete() {
	cb.context.Writer.Write([]byte("data: [DONE]\n\n"))
	cb.context.Writer.Flush()
	cb.done.Done()
}

func (cb *streamCallback) OnError(err error) {
	cb.context.Writer.Write([]byte(fmt.Sprintf("error: %v\n\n", err)))
	cb.context.Writer.Flush()
	cb.done.Done()
}

// convertToOpenAIFormat 将Anakin响应转换为OpenAI格式
func convertToOpenAIFormat(data string, model string) string {
	var anakinResp struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(data), &anakinResp); err != nil {
		return ""
	}

	resp := OpenAIResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Delta: &Delta{
					Content: anakinResp.Content,
				},
			},
		},
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return ""
	}
	return string(jsonResp)
}

// convertToOpenAIResponse 将Anakin响应转换为OpenAI响应对象
func (h *OpenAIHandler) convertToOpenAIResponse(anakinResponse, model string) OpenAIResponse {
	return OpenAIResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Usage: Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
		Choices: []Choice{
			{
				Index: 0,
				Message: &Message{
					Role:    "assistant",
					Content: anakinResponse,
				},
				FinishReason: "stop",
			},
		},
	}
}
