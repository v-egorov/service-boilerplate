package utils

import (
	"fmt"
	"os"
)

// HandleError handles CLI errors with consistent formatting
func HandleError(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %s\n", message)
		fmt.Fprintf(os.Stderr, "   Details: %v\n", err)
		os.Exit(1)
	}
}

// HandleAPIError handles API-related errors
func HandleAPIError(err error, operation string) error {
	if err != nil {
		return fmt.Errorf("API operation '%s' failed: %w", operation, err)
	}
	return nil
}

// ValidateRequired validates that required arguments are provided
func ValidateRequired(value, name string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ConfirmAction prompts user for confirmation
func ConfirmAction(message string) bool {
	fmt.Printf("⚠️  %s\n", message)
	fmt.Print("Continue? (y/N): ")

	var response string
	fmt.Scanln(&response)

	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
