package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

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

	// Create temp directory if not exists
	tempDir := "/tmp/audio_transcribe"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create temp directory",
		})
	}

	// Save file temporarily
	tempPath := filepath.Join(tempDir, file.Filename)
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
