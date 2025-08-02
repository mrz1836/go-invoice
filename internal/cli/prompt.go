package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Prompter provides interactive prompting capabilities with context support
type Prompter struct {
	reader *bufio.Reader
	logger Logger
}

// NewPrompter creates a new prompter instance
func NewPrompter(logger Logger) *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
		logger: logger,
	}
}

// PromptString prompts for a string value
func (p *Prompter) PromptString(ctx context.Context, prompt string, defaultValue string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	// Read input with context cancellation support
	type result struct {
		value string
		err   error
	}
	resultCh := make(chan result, 1)

	go func() {
		input, err := p.reader.ReadString('\n')
		if err != nil {
			resultCh <- result{err: fmt.Errorf("failed to read input: %w", err)}
			return
		}

		input = strings.TrimSpace(input)
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		resultCh <- result{value: input}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultCh:
		return res.value, res.err
	}
}

// PromptStringRequired prompts for a required string value
func (p *Prompter) PromptStringRequired(ctx context.Context, prompt string) (string, error) {
	for {
		value, err := p.PromptString(ctx, prompt, "")
		if err != nil {
			return "", err
		}

		if strings.TrimSpace(value) != "" {
			return value, nil
		}

		fmt.Println("❌ This field is required. Please provide a value.")
	}
}

// PromptInt prompts for an integer value
func (p *Prompter) PromptInt(ctx context.Context, prompt string, defaultValue int) (int, error) {
	defaultStr := ""
	if defaultValue != 0 {
		defaultStr = strconv.Itoa(defaultValue)
	}

	for {
		input, err := p.PromptString(ctx, prompt, defaultStr)
		if err != nil {
			return 0, err
		}

		if input == "" && defaultValue != 0 {
			return defaultValue, nil
		}

		value, err := strconv.Atoi(input)
		if err == nil {
			return value, nil
		}

		fmt.Printf("❌ Invalid number. Please enter a valid integer.\n")
	}
}

// PromptFloat prompts for a float value
func (p *Prompter) PromptFloat(ctx context.Context, prompt string, defaultValue float64) (float64, error) {
	defaultStr := ""
	if defaultValue != 0 {
		defaultStr = fmt.Sprintf("%.2f", defaultValue)
	}

	for {
		input, err := p.PromptString(ctx, prompt, defaultStr)
		if err != nil {
			return 0, err
		}

		if input == "" && defaultValue != 0 {
			return defaultValue, nil
		}

		value, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return value, nil
		}

		fmt.Printf("❌ Invalid number. Please enter a valid decimal number.\n")
	}
}

// PromptBool prompts for a yes/no answer
func (p *Prompter) PromptBool(ctx context.Context, prompt string, defaultValue bool) (bool, error) {
	defaultStr := "no"
	if defaultValue {
		defaultStr = "yes"
	}

	for {
		input, err := p.PromptString(ctx, prompt+" (yes/no)", defaultStr)
		if err != nil {
			return false, err
		}

		input = strings.ToLower(input)
		switch input {
		case "yes", "y", "true", "1":
			return true, nil
		case "no", "n", "false", "0":
			return false, nil
		default:
			fmt.Printf("❌ Please answer 'yes' or 'no'.\n")
		}
	}
}

// PromptSelect prompts the user to select from a list of options
func (p *Prompter) PromptSelect(ctx context.Context, prompt string, options []string, defaultIndex int) (int, string, error) {
	if len(options) == 0 {
		return -1, "", fmt.Errorf("no options provided")
	}

	fmt.Println(prompt)
	for i, option := range options {
		if i == defaultIndex {
			fmt.Printf("  %d. %s (default)\n", i+1, option)
		} else {
			fmt.Printf("  %d. %s\n", i+1, option)
		}
	}

	defaultStr := ""
	if defaultIndex >= 0 && defaultIndex < len(options) {
		defaultStr = strconv.Itoa(defaultIndex + 1)
	}

	for {
		input, err := p.PromptString(ctx, "Select an option", defaultStr)
		if err != nil {
			return -1, "", err
		}

		index, err := strconv.Atoi(input)
		if err == nil && index > 0 && index <= len(options) {
			return index - 1, options[index-1], nil
		}

		fmt.Printf("❌ Please select a valid option (1-%d).\n", len(options))
	}
}

// PromptMultiSelect prompts the user to select multiple options
func (p *Prompter) PromptMultiSelect(ctx context.Context, prompt string, options []string) ([]int, []string, error) {
	if len(options) == 0 {
		return nil, nil, fmt.Errorf("no options provided")
	}

	fmt.Println(prompt)
	fmt.Println("(Enter numbers separated by commas, or 'all' for all options)")
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	input, err := p.PromptString(ctx, "Select options", "")
	if err != nil {
		return nil, nil, err
	}

	if strings.ToLower(input) == "all" {
		indices := make([]int, len(options))
		for i := range indices {
			indices[i] = i
		}
		return indices, options, nil
	}

	var indices []int
	var selected []string

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		index, err := strconv.Atoi(part)
		if err == nil && index > 0 && index <= len(options) {
			indices = append(indices, index-1)
			selected = append(selected, options[index-1])
		}
	}

	if len(indices) == 0 {
		return nil, nil, fmt.Errorf("no valid options selected")
	}

	return indices, selected, nil
}

// PromptDate prompts for a date value
func (p *Prompter) PromptDate(ctx context.Context, prompt string, defaultValue time.Time) (time.Time, error) {
	defaultStr := ""
	if !defaultValue.IsZero() {
		defaultStr = defaultValue.Format("2006-01-02")
	}

	for {
		input, err := p.PromptString(ctx, prompt+" (YYYY-MM-DD)", defaultStr)
		if err != nil {
			return time.Time{}, err
		}

		if input == "" && !defaultValue.IsZero() {
			return defaultValue, nil
		}

		// Parse the date
		date, err := time.Parse("2006-01-02", input)
		if err == nil {
			return date, nil
		}

		// Try parsing with common formats
		formats := []string{
			"01/02/2006",
			"2006/01/02",
			"02-01-2006",
			"Jan 2, 2006",
			"January 2, 2006",
		}

		for _, format := range formats {
			if date, err := time.Parse(format, input); err == nil {
				return date, nil
			}
		}

		fmt.Printf("❌ Invalid date format. Please use YYYY-MM-DD format (e.g., 2024-01-15).\n")
	}
}

// PromptConfirm prompts for confirmation with a custom message
func (p *Prompter) PromptConfirm(ctx context.Context, message string) (bool, error) {
	fmt.Println(message)
	return p.PromptBool(ctx, "Continue?", false)
}

// PromptPassword prompts for a password (input is hidden)
func (p *Prompter) PromptPassword(ctx context.Context, prompt string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	fmt.Printf("%s: ", prompt)

	// Note: In a real implementation, we would use a library like golang.org/x/term
	// to hide the password input. For now, we'll use regular input with a warning.
	fmt.Println("(WARNING: Password will be visible)")

	return p.PromptString(ctx, "", "")
}

// Logger interface for logging operations
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
}
