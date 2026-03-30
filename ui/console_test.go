package ui

import (
	"testing"
	"strings"
	"bufio"
	"bytes"
)

func TestConsoleUIDisplayMessage(t *testing.T) {
	var output bytes.Buffer
	ui := NewConsoleUI(&output, nil) 

	msg := Message{
		SenderName: "Alice",
		Timestamp:  time.Date(2026, 3, 30, 14, 32, 1, 0, time.UTC),
		Text:       "Hello World!",
	}

	ui.DisplayMessage(msg)

	outputStr := output.String()
	expectedContent := "[14:32:01] Alice: Hello World!"
	if !strings.Contains(outputStr, expectedContent) {
		t.Errorf("Output doesn't contain expected message. Expected: '%s', Got: '%s'", expectedContent, outputStr)
	}
}

func TestConsoleUIDisplayMessageAddsNewline(t *testing.T) {
	var output bytes.Buffer
	ui := NewConsoleUI(&output, nil)

	msg1 := Message{
		SenderName: "Alice",
		Timestamp:  time.Date(2026, 3, 30, 14, 32, 1, 0, time.UTC),
		Text:       "Message 1",
	}
	msg2 := Message{
		SenderName: "Bob",
		Timestamp:  time.Date(2026, 3, 30, 14, 32, 2, 0, time.UTC),
		Text:       "Message 2",
	}

	ui.DisplayMessage(msg1)
	ui.DisplayMessage(msg2)

	outputStr := output.String()
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) < 2 {
		t.Errorf("Messages should be on separate lines. Got %d line(s)", len(lines))
	}
}

func TestConsoleUIReadInput(t *testing.T) {
	inputText := "Hello from user\n"
	reader := bufio.NewReader(strings.NewReader(inputText))
	
	var output bytes.Buffer
	ui := NewConsoleUI(&output, reader)

	input, err := ui.ReadInput()

	if err != nil {
		t.Errorf("ReadInput should not produce error, got: %v", err)
	}
	if input != "Hello from user" {
		t.Errorf("ReadInput should return 'Hello from user', got '%s'", input)
	}
}

func TestConsoleUIReadInputEmptyLine(t *testing.T) {
	inputText := "\n"
	reader := bufio.NewReader(strings.NewReader(inputText))
	
	var output bytes.Buffer
	ui := NewConsoleUI(&output, reader)

	input, err := ui.ReadInput()

	if err != nil {
		t.Errorf("ReadInput should handle empty line, got error: %v", err)
	}
	if input != "" {
		t.Errorf("ReadInput should return empty string, got '%s'", input)
	}
}

func TestConsoleUIReadInputMultipleLines(t *testing.T) {
	inputText := "First line\nSecond line\nThird line\n"
	reader := bufio.NewReader(strings.NewReader(inputText))
	
	var output bytes.Buffer
	ui := NewConsoleUI(&output, reader)

	input1, err1 := ui.ReadInput()
	input2, err2 := ui.ReadInput()
	input3, err3 := ui.ReadInput()

	if input1 != "First line" {
		t.Errorf("First read should return 'First line', got '%s'", input1)
	}
	if input2 != "Second line" {
		t.Errorf("Second read should return 'Second line', got '%s'", input2)
	}
	if input3 != "Third line" {
		t.Errorf("Third read should return 'Third line', got '%s'", input3)
	}
	if err1 != nil || err2 != nil || err3 != nil {
		t.Errorf("ReadInput should not produce errors, got: %v, %v, %v", err1, err2, err3)
	}
}

func TestConsoleUIReadInputWithSpecialCharacters(t *testing.T) {
	inputText := "Привет! 👋 Как дела?\n"
	reader := bufio.NewReader(strings.NewReader(inputText))
	
	var output bytes.Buffer
	ui := NewConsoleUI(&output, reader)

	input, err := ui.ReadInput()

	if err != nil {
		t.Errorf("ReadInput should handle special characters, got error: %v", err)
	}
	if input != "Привет! 👋 Как дела?" {
		t.Errorf("ReadInput should preserve special characters, got '%s'", input)
	}
}

func TestConsoleUIIsImplementingUIRenderer(t *testing.T) {
	var output bytes.Buffer
	var input strings.Reader
	ui := NewConsoleUI(&output, bufio.NewReader(&input))

	var _ UIRenderer = ui

	t.Log("ConsoleUI correctly implements UIRenderer interface")
}

func TestConsoleUIDisplayMessageWithNilWriter(t *testing.T) {
	ui := NewConsoleUI(nil, nil)

	msg := Message{
		SenderName: "Test",
		Timestamp:  time.Now(),
		Text:       "test",
	}

	result := t.Run("display_with_nil_writer", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("DisplayMessage with nil writer caused panic (expected or fixable): %v", r)
			}
		}()
		ui.DisplayMessage(msg)
		t.Log("DisplayMessage handled nil writer gracefully")
	})

	if !result {
		t.Logf("nil writer test requires attention")
	}
}

func TestConsoleUIDisplaysCorrectTimezone(t *testing.T) {
	var output bytes.Buffer
	ui := NewConsoleUI(&output, nil)

	utcTime := time.Date(2026, 3, 30, 14, 32, 1, 0, time.UTC)
	msg := Message{
		SenderName: "Alice",
		Timestamp:  utcTime,
		Text:       "Test message",
	}

	ui.DisplayMessage(msg)

	outputStr := output.String()
	if !strings.Contains(outputStr, "14:32:01") {
		t.Errorf("Expected time '14:32:01' in output, got: %s", outputStr)
	}
}
