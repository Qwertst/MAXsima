package model

import (
	"testing"
	"time"
)

func TestMessageFormat(t *testing.T) {
	fixedTime := time.Date(2026, 3, 30, 14, 32, 1, 0, time.UTC)

	msg := Message{
		SenderName: "Alice",
		Timestamp:  fixedTime,
		Text:       "Hello World!",
	}

	formatted := msg.Format()

	expectedFormat := "[2026-03-30 14:32:01] Alice: Hello World!"
	if formatted != expectedFormat {
		t.Errorf("Expected '%s', got '%s'", expectedFormat, formatted)
	}
}

func TestMessageFormatWithEmptyText(t *testing.T) {
	fixedTime := time.Date(2026, 3, 30, 15, 45, 30, 0, time.UTC)
	msg := Message{
		SenderName: "Bob",
		Timestamp:  fixedTime,
		Text:       "",
	}

	formatted := msg.Format()

	expectedFormat := "[2026-03-30 15:45:30] Bob: "
	if formatted != expectedFormat {
		t.Errorf("Expected '%s', got '%s'", expectedFormat, formatted)
	}
}

func TestMessageFormatWithSpecialCharacters(t *testing.T) {
	fixedTime := time.Date(2026, 3, 30, 16, 20, 15, 0, time.UTC)
	msg := Message{
		SenderName: "Алиса",
		Timestamp:  fixedTime,
		Text:       "Привет! 👋 Как дела?",
	}

	formatted := msg.Format()

	expectedFormat := "[2026-03-30 16:20:15] Алиса: Привет! 👋 Как дела?"
	if formatted != expectedFormat {
		t.Errorf("Expected '%s', got '%s'", expectedFormat, formatted)
	}
}

func TestMessageFormatWithLongText(t *testing.T) {
	fixedTime := time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC)
	longText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."

	msg := Message{
		SenderName: "Charlie",
		Timestamp:  fixedTime,
		Text:       longText,
	}

	formatted := msg.Format()

	if !contains(formatted, longText) {
		t.Errorf("Formatted message doesn't contain full text")
	}
	if !contains(formatted, "[2026-03-30 17:00:00] Charlie:") {
		t.Errorf("Formatted message doesn't contain proper header")
	}
}

func TestMessageCreation(t *testing.T) {
	senderName := "David"
	timestamp := time.Date(2026, 3, 30, 18, 15, 45, 0, time.UTC)
	text := "Test message"

	msg := Message{
		SenderName: senderName,
		Timestamp:  timestamp,
		Text:       text,
	}

	if msg.SenderName != senderName {
		t.Errorf("SenderName not set correctly: expected '%s', got '%s'", senderName, msg.SenderName)
	}
	if msg.Timestamp != timestamp {
		t.Errorf("Timestamp not set correctly: expected %v, got %v", timestamp, msg.Timestamp)
	}
	if msg.Text != text {
		t.Errorf("Text not set correctly: expected '%s', got '%s'", text, msg.Text)
	}
}

func TestMessageTimestampPrecision(t *testing.T) {
	originalTime := time.Date(2026, 3, 30, 19, 30, 0, 0, time.UTC)
	msg := Message{
		SenderName: "Eve",
		Timestamp:  originalTime,
		Text:       "Precision test",
	}

	retrievedTime := msg.Timestamp

	if !retrievedTime.Equal(originalTime) {
		t.Errorf("Timestamp changed: expected %v, got %v", originalTime, retrievedTime)
	}
}

func contains(str, substring string) bool {
	return len(str) >= len(substring) && str[:len(substring)] == substring ||
		(len(str) > len(substring) && str[len(str)-len(substring):] == substring) ||
		indexOf(str, substring) >= 0
}

func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
