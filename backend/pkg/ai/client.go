package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Client handles AI operations
type Client struct {
	openaiKey      string
	anthropicKey   string
	yandexAPIKey   string
	yandexFolderID string
	httpClient     *http.Client
}

// NewClient creates a new AI client
func NewClient(openaiKey, anthropicKey, yandexAPIKey, yandexFolderID string) *Client {
	return &Client{
		openaiKey:      openaiKey,
		anthropicKey:   anthropicKey,
		yandexAPIKey:   yandexAPIKey,
		yandexFolderID: yandexFolderID,
		httpClient:     &http.Client{},
	}
}

// AnalysisContext holds context for meeting analysis
type AnalysisContext struct {
	Transcript        string
	EmployeeContext   string
	MeetingsHistory   string
	ProjectContext    string
	Participants      string
	WhisperTranscript string
	YandexTranscript  string
}

// TranscribeWhisper transcribes audio using OpenAI Whisper
func (c *Client) TranscribeWhisper(filePath string) (string, error) {
	if c.openaiKey == "" {
		return "", nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "ru")
	writer.WriteField("response_format", "text")
	writer.WriteField("prompt", "Это рабочая встреча. Обсуждаются проекты, задачи, KPI, спринты, дедлайны.")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.openaiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("whisper error %d: %s", resp.StatusCode, string(result))
	}

	return string(result), nil
}

// TranscribeYandex transcribes audio using Yandex SpeechKit
func (c *Client) TranscribeYandex(filePath string) (string, error) {
	if c.yandexAPIKey == "" || c.yandexFolderID == "" {
		return "", nil
	}

	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	audioFormat := "oggopus"
	if ext == ".mp3" {
		audioFormat = "mp3"
	} else if ext == ".wav" {
		audioFormat = "lpcm"
	}

	url := fmt.Sprintf("https://stt.api.cloud.yandex.net/speech/v1/stt:recognize?folderId=%s&lang=ru-RU&format=%s&sampleRateHertz=48000",
		c.yandexFolderID, audioFormat)

	req, err := http.NewRequest("POST", url, bytes.NewReader(audioData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Api-Key "+c.yandexAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("yandex STT error %d: %s", resp.StatusCode, string(result))
	}

	var response struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return "", err
	}

	return response.Result, nil
}

// MergeTranscripts merges two transcripts using Claude
func (c *Client) MergeTranscripts(whisper, yandex string) (string, error) {
	if c.anthropicKey == "" {
		if whisper != "" {
			return whisper, nil
		}
		return yandex, nil
	}

	if whisper == "" {
		return yandex, nil
	}
	if yandex == "" {
		return whisper, nil
	}

	ctx := AnalysisContext{
		WhisperTranscript: whisper,
		YandexTranscript:  yandex,
	}

	prompt, err := c.renderPrompt(TranscriptMergePrompt, ctx)
	if err != nil {
		return whisper, err
	}

	return c.callClaude(prompt)
}

// Analyze analyzes a transcript using the appropriate prompt
func (c *Client) Analyze(categoryCode string, ctx AnalysisContext) (map[string]interface{}, error) {
	promptTemplate, ok := Prompts[categoryCode]
	if !ok {
		promptTemplate = Prompts["default"]
	}

	prompt, err := c.renderPrompt(promptTemplate, ctx)
	if err != nil {
		return nil, err
	}

	responseText, err := c.callClaude(prompt)
	if err != nil {
		return nil, err
	}

	// Extract JSON from response
	responseText = extractJSON(responseText)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %v, response: %s", err, responseText[:min(500, len(responseText))])
	}

	return result, nil
}

func (c *Client) renderPrompt(promptTemplate string, ctx AnalysisContext) (string, error) {
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *Client) callClaude(prompt string) (string, error) {
	if c.anthropicKey == "" {
		return "", fmt.Errorf("Anthropic API key not configured")
	}

	payload := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 8000,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Claude error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return response.Content[0].Text, nil
}

func extractJSON(text string) string {
	// Try to extract JSON from markdown code blocks
	if strings.Contains(text, "```json") {
		parts := strings.Split(text, "```json")
		if len(parts) > 1 {
			jsonPart := strings.Split(parts[1], "```")[0]
			return strings.TrimSpace(jsonPart)
		}
	}

	if strings.Contains(text, "```") {
		parts := strings.Split(text, "```")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}

	return strings.TrimSpace(text)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
