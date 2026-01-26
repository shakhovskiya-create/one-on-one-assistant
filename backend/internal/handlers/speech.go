package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Audio upload security constants
const (
	// MaxAudioFileSize is the maximum allowed audio file size (100 MB for long recordings)
	MaxAudioFileSize = 100 * 1024 * 1024
)

// AllowedAudioTypes defines the whitelist of allowed audio extensions and MIME types
var AllowedAudioTypes = map[string][]string{
	".mp3":  {"audio/mpeg"},
	".wav":  {"audio/wav", "audio/x-wav", "audio/wave"},
	".ogg":  {"audio/ogg", "application/ogg"},
	".webm": {"audio/webm", "video/webm"},
	".m4a":  {"audio/mp4", "audio/x-m4a"},
	".flac": {"audio/flac", "audio/x-flac"},
}

// validateAudioUpload performs security validation on uploaded audio file
func validateAudioUpload(filename string, size int64, contentBuffer []byte) error {
	// 1. Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("audio file must have an extension")
	}

	allowedMimes, ok := AllowedAudioTypes[ext]
	if !ok {
		return fmt.Errorf("audio type not allowed: %s (allowed: .mp3, .wav, .ogg, .webm, .m4a, .flac)", ext)
	}

	// 2. Check file size
	if size > MaxAudioFileSize {
		return fmt.Errorf("audio file too large: %d bytes (max: %d bytes)", size, MaxAudioFileSize)
	}

	// 3. Detect MIME type from magic bytes
	detectedMime := http.DetectContentType(contentBuffer)
	detectedMime = strings.Split(detectedMime, ";")[0]
	detectedMime = strings.TrimSpace(detectedMime)

	// 4. Validate MIME type
	mimeValid := false
	for _, allowed := range allowedMimes {
		if detectedMime == allowed {
			mimeValid = true
			break
		}
	}

	// Some audio formats may be detected as application/octet-stream
	if !mimeValid && detectedMime == "application/octet-stream" {
		mimeValid = true // Allow, as audio formats may not have clear magic bytes
	}

	// Video/webm container can hold audio
	if !mimeValid && ext == ".webm" && strings.HasPrefix(detectedMime, "video/") {
		mimeValid = true
	}

	if !mimeValid {
		return fmt.Errorf("audio content does not match extension: detected %s for %s file", detectedMime, ext)
	}

	return nil
}

// TranscribeAudio transcribes uploaded audio using Whisper or Yandex SpeechKit
func (h *Handler) TranscribeAudio(c *fiber.Ctx) error {
	if h.AI == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "AI client not configured",
		})
	}

	// Get the uploaded file
	file, err := c.FormFile("audio")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No audio file provided",
		})
	}

	// 1. Early size check
	if file.Size > MaxAudioFileSize {
		return c.Status(413).JSON(fiber.Map{
			"error": fmt.Sprintf("Audio file too large: %d bytes (max: %d bytes)", file.Size, MaxAudioFileSize),
		})
	}

	// 2. Open and read header for validation
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to read audio file",
		})
	}
	defer fileContent.Close()

	// Read first 512 bytes for MIME detection
	sniffBuffer := make([]byte, 512)
	n, err := fileContent.Read(sniffBuffer)
	if err != nil && err != io.EOF {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to read audio file header",
		})
	}
	sniffBuffer = sniffBuffer[:n]

	// 3. Validate audio file
	if err := validateAudioUpload(file.Filename, file.Size, sniffBuffer); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create temp directory if not exists
	tempDir := "/tmp/audio_transcribe"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create temp directory",
		})
	}

	// Generate safe filename to prevent path traversal
	ext := strings.ToLower(filepath.Ext(file.Filename))
	safeFilename := uuid.New().String() + ext
	tempPath := filepath.Join(tempDir, safeFilename)

	// Save file temporarily
	if err := c.SaveFile(file, tempPath); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save uploaded file",
		})
	}
	defer os.Remove(tempPath)

	// Determine transcription method based on query param
	service := c.Query("service", "auto")

	var transcript string
	var whisperResult, yandexResult string
	var transcribeErr error

	switch service {
	case "whisper":
		transcript, transcribeErr = h.AI.TranscribeWhisper(tempPath)
	case "yandex":
		transcript, transcribeErr = h.AI.TranscribeYandex(tempPath)
	case "auto", "both":
		// Try both and merge if possible
		whisperResult, _ = h.AI.TranscribeWhisper(tempPath)
		yandexResult, _ = h.AI.TranscribeYandex(tempPath)

		if whisperResult != "" && yandexResult != "" {
			transcript, transcribeErr = h.AI.MergeTranscripts(whisperResult, yandexResult)
		} else if whisperResult != "" {
			transcript = whisperResult
		} else if yandexResult != "" {
			transcript = yandexResult
		} else {
			transcribeErr = fmt.Errorf("no transcription service available")
		}
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid service. Use: whisper, yandex, auto, or both",
		})
	}

	if transcribeErr != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Transcription failed",
			"details": transcribeErr.Error(),
		})
	}

	if transcript == "" {
		return c.Status(500).JSON(fiber.Map{
			"error": "No transcription result. Check API keys configuration.",
		})
	}

	return c.JSON(fiber.Map{
		"transcript": transcript,
		"whisper":    whisperResult,
		"yandex":     yandexResult,
	})
}
