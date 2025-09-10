package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PDFProcessModel represents the PDF processing state
type PDFProcessModel struct {
	selectedFile    string
	extractedText   string
	step           int // 0: extract, 1: configure, 2: generate
	errorMsg       string
	successMsg     string
	loading        bool
	
	// Configuration
	numQuestions   string
	questionTypes  map[string]bool
	testName       string
	testDesc       string
	
	// Input mode
	inputMode      string // "num_questions", "test_name", "test_desc", ""
	input          string
	cursor         int
}

// NewPDFProcessModel creates a new PDF process model
func NewPDFProcessModel() *PDFProcessModel {
	return &PDFProcessModel{
		step: 0,
		numQuestions: "5",
		questionTypes: map[string]bool{
			"multiple_choice": true,
			"true_false":     false,
			"short_answer":   false,
		},
		testName: "Generated Test",
		testDesc: "Test generated from PDF",
	}
}

// updatePDFProcess handles PDF processing updates
func (a *App) updatePDFProcess(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.pdfProcess.loading {
		return a, nil // Ignore input while loading
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.pdfProcess.inputMode != "" {
			return a.handlePDFInputMode(msg)
		}
		
		switch a.pdfProcess.step {
		case 0: // Extract step
			switch msg.String() {
			case "enter", " ":
				return a.extractPDFText()
			}
		case 1: // Configure step
			return a.handleConfigureStep(msg)
		case 2: // Generate step
			switch msg.String() {
			case "enter", " ":
				return a.generateQuestions()
			case "b":
				a.pdfProcess.step = 1
			}
		}
	}
	return a, nil
}

// viewPDFProcess renders the PDF processing view
func (a *App) viewPDFProcess() string {
	s := a.renderHeader("PDF Question Generation")
	
	if a.pdfProcess.errorMsg != "" {
		s += a.renderError(a.pdfProcess.errorMsg)
		a.pdfProcess.errorMsg = ""
	}
	
	if a.pdfProcess.successMsg != "" {
		s += a.renderSuccess(a.pdfProcess.successMsg)
		a.pdfProcess.successMsg = ""
	}
	
	if a.pdfProcess.loading {
		s += "⏳ Processing... Please wait...\n\n"
		return s + a.renderFooter()
	}
	
	switch a.pdfProcess.step {
	case 0:
		return s + a.viewExtractStep() + a.renderFooter()
	case 1:
		return s + a.viewConfigureStep() + a.renderFooter()
	case 2:
		return s + a.viewGenerateStep() + a.renderFooter()
	default:
		return s + "Unknown step" + a.renderFooter()
	}
}

// viewExtractStep renders the text extraction step
func (a *App) viewExtractStep() string {
	s := fmt.Sprintf("Selected PDF: %s\n\n", a.pdfProcess.selectedFile)
	
	if a.pdfProcess.extractedText == "" {
		s += "Press Enter to extract text from the PDF\n"
	} else {
		s += "✅ Text extracted successfully!\n\n"
		s += "Preview:\n"
		preview := a.pdfProcessor.GetTextSummary(a.pdfProcess.extractedText, 200)
		s += borderStyle.Render(preview) + "\n\n"
		s += "Press Enter to continue to configuration\n"
	}
	
	return s
}

// viewConfigureStep renders the configuration step
func (a *App) viewConfigureStep() string {
	s := "Configure Question Generation:\n\n"
	
	if a.pdfProcess.inputMode != "" {
		return s + a.viewInputMode()
	}
	
	// Number of questions
	cursor := " "
	if a.pdfProcess.cursor == 0 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Number of questions: %s (press 'n' to edit)\n", cursor, a.pdfProcess.numQuestions)
	
	// Question types
	cursor = " "
	if a.pdfProcess.cursor == 1 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Question types:\n", cursor)
	for qType, enabled := range a.pdfProcess.questionTypes {
		status := "❌"
		if enabled {
			status = "✅"
		}
		s += fmt.Sprintf("   %s %s\n", status, a.getQuestionTypeDisplay(qType))
	}
	s += "   (press 't' to toggle types)\n\n"
	
	// Test name
	cursor = " "
	if a.pdfProcess.cursor == 2 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Test name: %s (press 'e' to edit)\n", cursor, a.pdfProcess.testName)
	
	// Test description
	cursor = " "
	if a.pdfProcess.cursor == 3 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Test description: %s (press 'd' to edit)\n\n", cursor, a.pdfProcess.testDesc)
	
	s += "Press Enter to generate questions, arrow keys to navigate\n"
	
	return s
}

// viewGenerateStep renders the generation step
func (a *App) viewGenerateStep() string {
	s := "Ready to Generate Questions:\n\n"
	s += fmt.Sprintf("📄 PDF: %s\n", a.pdfProcess.selectedFile)
	s += fmt.Sprintf("📝 Test: %s\n", a.pdfProcess.testName)
	s += fmt.Sprintf("🔢 Questions: %s\n", a.pdfProcess.numQuestions)
	
	var enabledTypes []string
	for qType, enabled := range a.pdfProcess.questionTypes {
		if enabled {
			enabledTypes = append(enabledTypes, a.getQuestionTypeDisplay(qType))
		}
	}
	s += fmt.Sprintf("📋 Types: %s\n\n", strings.Join(enabledTypes, ", "))
	
	s += "Press Enter to generate questions, 'b' to go back\n"
	
	return s
}

// viewInputMode renders input mode interface
func (a *App) viewInputMode() string {
	var prompt string
	switch a.pdfProcess.inputMode {
	case "num_questions":
		prompt = "Enter number of questions:"
	case "test_name":
		prompt = "Enter test name:"
	case "test_desc":
		prompt = "Enter test description:"
	}
	
	s := prompt + "\n"
	s += "> " + a.pdfProcess.input + "\n\n"
	s += "Press Enter to confirm, Esc to cancel\n"
	
	return s
}

// handleConfigureStep handles configuration step input
func (a *App) handleConfigureStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if a.pdfProcess.cursor > 0 {
			a.pdfProcess.cursor--
		}
	case "down", "j":
		if a.pdfProcess.cursor < 3 {
			a.pdfProcess.cursor++
		}
	case "n":
		if a.pdfProcess.cursor == 0 {
			a.pdfProcess.inputMode = "num_questions"
			a.pdfProcess.input = a.pdfProcess.numQuestions
		}
	case "t":
		if a.pdfProcess.cursor == 1 {
			return a.toggleQuestionTypes()
		}
	case "e":
		if a.pdfProcess.cursor == 2 {
			a.pdfProcess.inputMode = "test_name"
			a.pdfProcess.input = a.pdfProcess.testName
		}
	case "d":
		if a.pdfProcess.cursor == 3 {
			a.pdfProcess.inputMode = "test_desc"
			a.pdfProcess.input = a.pdfProcess.testDesc
		}
	case "enter", " ":
		a.pdfProcess.step = 2
	}
	return a, nil
}

// handlePDFInputMode handles input mode for PDF processing
func (a *App) handlePDFInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm input
		switch a.pdfProcess.inputMode {
		case "num_questions":
			if num, err := strconv.Atoi(strings.TrimSpace(a.pdfProcess.input)); err == nil && num > 0 && num <= 50 {
				a.pdfProcess.numQuestions = a.pdfProcess.input
			} else {
				a.pdfProcess.errorMsg = "Please enter a valid number between 1 and 50"
			}
		case "test_name":
			if err := a.validateInput(a.pdfProcess.input, 1); err == nil {
				a.pdfProcess.testName = strings.TrimSpace(a.pdfProcess.input)
			} else {
				a.pdfProcess.errorMsg = err.Error()
			}
		case "test_desc":
			a.pdfProcess.testDesc = strings.TrimSpace(a.pdfProcess.input)
		}
		a.pdfProcess.inputMode = ""
		a.pdfProcess.input = ""
	case "esc":
		// Cancel input
		a.pdfProcess.inputMode = ""
		a.pdfProcess.input = ""
	case "backspace":
		// Remove last character
		if len(a.pdfProcess.input) > 0 {
			a.pdfProcess.input = a.pdfProcess.input[:len(a.pdfProcess.input)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			a.pdfProcess.input += msg.String()
		}
	}
	return a, nil
}

// extractPDFText extracts text from the selected PDF
func (a *App) extractPDFText() (tea.Model, tea.Cmd) {
	if a.pdfProcess.extractedText != "" {
		a.pdfProcess.step = 1
		return a, nil
	}
	
	a.pdfProcess.loading = true
	
	// Extract text from PDF
	text, err := a.pdfProcessor.ExtractText(a.pdfProcess.selectedFile)
	if err != nil {
		a.pdfProcess.errorMsg = fmt.Sprintf("Failed to extract text: %v", err)
		a.pdfProcess.loading = false
		return a, nil
	}
	
	a.pdfProcess.extractedText = text
	a.pdfProcess.successMsg = "Text extracted successfully!"
	a.pdfProcess.loading = false
	a.pdfProcess.step = 1
	
	return a, nil
}

// generateQuestions generates questions using ChatGPT
func (a *App) generateQuestions() (tea.Model, tea.Cmd) {
	a.pdfProcess.loading = true
	
	// Get enabled question types
	var questionTypes []string
	for qType, enabled := range a.pdfProcess.questionTypes {
		if enabled {
			questionTypes = append(questionTypes, qType)
		}
	}
	
	if len(questionTypes) == 0 {
		a.pdfProcess.errorMsg = "Please select at least one question type"
		a.pdfProcess.loading = false
		a.pdfProcess.step = 1
		return a, nil
	}
	
	numQuestions, _ := strconv.Atoi(a.pdfProcess.numQuestions)
	
	// Generate questions using ChatGPT
	generatedQuestions, err := a.chatGPT.GenerateQuestions(a.pdfProcess.extractedText, numQuestions, questionTypes)
	if err != nil {
		a.pdfProcess.errorMsg = fmt.Sprintf("Failed to generate questions: %v", err)
		a.pdfProcess.loading = false
		return a, nil
	}
	
	// Create test in database
	test, err := a.db.CreateTest(a.pdfProcess.testName, a.pdfProcess.testDesc)
	if err != nil {
		a.pdfProcess.errorMsg = fmt.Sprintf("Failed to create test: %v", err)
		a.pdfProcess.loading = false
		return a, nil
	}
	
	// Save questions to database
	for _, gq := range generatedQuestions {
		_, err := a.db.CreateQuestion(test.ID, gq.Question, gq.Type, gq.CorrectAnswer, gq.Explanation, gq.Options)
		if err != nil {
			a.pdfProcess.errorMsg = fmt.Sprintf("Failed to save question: %v", err)
			a.pdfProcess.loading = false
			return a, nil
		}
	}
	
	a.pdfProcess.loading = false
	a.pdfProcess.successMsg = fmt.Sprintf("Successfully generated %d questions!", len(generatedQuestions))
	
	// Switch to main menu after success
	a.currentView = MainMenuView
	
	return a, nil
}

// toggleQuestionTypes toggles question type selection
func (a *App) toggleQuestionTypes() (tea.Model, tea.Cmd) {
	// Simple toggle - cycle through enabling different types
	types := []string{"multiple_choice", "true_false", "short_answer"}
	
	// Find currently enabled type and move to next
	for i, qType := range types {
		if a.pdfProcess.questionTypes[qType] {
			a.pdfProcess.questionTypes[qType] = false
			nextType := types[(i+1)%len(types)]
			a.pdfProcess.questionTypes[nextType] = true
			break
		}
	}
	
	return a, nil
}