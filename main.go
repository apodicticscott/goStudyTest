package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	tea "github.com/charmbracelet/bubbletea"
	"pdf-test-generator/tui"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using system environment variables.")
	}

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || apiKey == "your_openai_api_key_here" {
		log.Println("Warning: OPENAI_API_KEY not set or using placeholder. ChatGPT features will be disabled.")
		apiKey = ""
	}

	// Initialize TUI application
	app, err := tui.NewApp("test_generator.db", apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}