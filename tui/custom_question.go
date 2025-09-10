package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CustomQuestionModel represents the custom question creation state
type CustomQuestionModel struct {
	step           int    // 0: test info, 1: question creation, 2: review
	cursor         int
	inputMode      string // "test_name", "test_desc", "question", "answer", "explanation", "option"
	input          string
	errorMsg       string
	successMsg     string
	
	// Test info
	testName       string
	testDesc       string
	
	// Current question being created
	currentQuestion struct {
		text        string
		qType       string
		options     []string
		correctAnswer string
		explanation string
	}
	
	// Questions created so far
	questions      []QuestionData
	questionTypes  []string
	typeIndex      int
	optionIndex    int
}

// QuestionData represents a created question
type QuestionData struct {
	Text          string
	Type          string
	Options       []string
	CorrectAnswer string
	Explanation   string
}

// NewCustomQuestionModel creates a new custom question model
func NewCustomQuestionModel() *CustomQuestionModel {
	return &CustomQuestionModel{
		step: 0,
		testName: "Custom Test",
		testDesc: "Custom created test",
		questionTypes: []string{"multiple_choice", "true_false", "short_answer"},
		currentQuestion: struct {
			text        string
			qType       string
			options     []string
			correctAnswer string
			explanation string
		}{
			qType: "multiple_choice",
			options: make([]string, 4), // Default 4 options for multiple choice
		},
	}
}

// updateCustomQuestion handles custom question updates
func (a *App) updateCustomQuestion(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.customQuestion.inputMode != "" {
			return a.handleCustomQuestionInput(msg)
		}
		
		switch a.customQuestion.step {
		case 0: // Test info step
			return a.handleTestInfoStep(msg)
		case 1: // Question creation step
			return a.handleQuestionCreationStep(msg)
		case 2: // Review step
			return a.handleReviewStep(msg)
		}
	}
	return a, nil
}

// viewCustomQuestion renders the custom question view
func (a *App) viewCustomQuestion() string {
	s := a.renderHeader("Create Custom Questions")
	
	if a.customQuestion.errorMsg != "" {
		s += a.renderError(a.customQuestion.errorMsg)
		a.customQuestion.errorMsg = ""
	}
	
	if a.customQuestion.successMsg != "" {
		s += a.renderSuccess(a.customQuestion.successMsg)
		a.customQuestion.successMsg = ""
	}
	
	switch a.customQuestion.step {
	case 0:
		return s + a.viewTestInfoStep() + a.renderFooter()
	case 1:
		return s + a.viewQuestionCreationStep() + a.renderFooter()
	case 2:
		return s + a.viewReviewStep() + a.renderFooter()
	default:
		return s + "Unknown step" + a.renderFooter()
	}
}

// viewTestInfoStep renders the test info step
func (a *App) viewTestInfoStep() string {
	s := "Step 1: Test Information\n\n"
	
	if a.customQuestion.inputMode != "" {
		return s + a.viewCustomQuestionInputMode()
	}
	
	// Test name
	cursor := " "
	if a.customQuestion.cursor == 0 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Test Name: %s (press 'n' to edit)\n", cursor, a.customQuestion.testName)
	
	// Test description
	cursor = " "
	if a.customQuestion.cursor == 1 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Test Description: %s (press 'd' to edit)\n\n", cursor, a.customQuestion.testDesc)
	
	s += "Press Enter to continue to question creation\n"
	s += "Use arrow keys to navigate, letters to edit\n"
	
	return s
}

// viewQuestionCreationStep renders the question creation step
func (a *App) viewQuestionCreationStep() string {
	s := fmt.Sprintf("Step 2: Create Questions (%d created so far)\n\n", len(a.customQuestion.questions))
	
	if a.customQuestion.inputMode != "" {
		return s + a.viewCustomQuestionInputMode()
	}
	
	// Question type selection
	cursor := " "
	if a.customQuestion.cursor == 0 {
		cursor = ">"
	}
	qType := a.getQuestionTypeDisplay(a.customQuestion.currentQuestion.qType)
	s += fmt.Sprintf("%s Question Type: %s (press 't' to change)\n", cursor, qType)
	
	// Question text
	cursor = " "
	if a.customQuestion.cursor == 1 {
		cursor = ">"
	}
	questionPreview := a.customQuestion.currentQuestion.text
	if len(questionPreview) > 50 {
		questionPreview = questionPreview[:50] + "..."
	}
	s += fmt.Sprintf("%s Question: %s (press 'q' to edit)\n", cursor, questionPreview)
	
	// Options (for multiple choice)
	if a.customQuestion.currentQuestion.qType == "multiple_choice" {
		cursor = " "
		if a.customQuestion.cursor == 2 {
			cursor = ">"
		}
		s += fmt.Sprintf("%s Options: (press 'o' to edit)\n", cursor)
		for i, option := range a.customQuestion.currentQuestion.options {
			optionText := option
			if optionText == "" {
				optionText = "[empty]"
			}
			s += fmt.Sprintf("   %c) %s\n", 'A'+i, optionText)
		}
	}
	
	// Correct answer
	cursor = " "
	if a.customQuestion.cursor == 3 {
		cursor = ">"
	}
	s += fmt.Sprintf("%s Correct Answer: %s (press 'a' to edit)\n", cursor, a.customQuestion.currentQuestion.correctAnswer)
	
	// Explanation
	cursor = " "
	if a.customQuestion.cursor == 4 {
		cursor = ">"
	}
	explanationPreview := a.customQuestion.currentQuestion.explanation
	if len(explanationPreview) > 50 {
		explanationPreview = explanationPreview[:50] + "..."
	}
	s += fmt.Sprintf("%s Explanation: %s (press 'e' to edit)\n\n", cursor, explanationPreview)
	
	s += "Press 's' to save this question and create another\n"
	s += "Press 'f' to finish and review all questions\n"
	s += "Use arrow keys to navigate\n"
	
	return s
}

// viewReviewStep renders the review step
func (a *App) viewReviewStep() string {
	s := fmt.Sprintf("Step 3: Review Questions (%d total)\n\n", len(a.customQuestion.questions))
	
	if len(a.customQuestion.questions) == 0 {
		s += "No questions created yet. Go back to create some questions.\n\n"
		s += "Press 'b' to go back\n"
		return s
	}
	
	s += fmt.Sprintf("Test: %s\n", a.customQuestion.testName)
	s += fmt.Sprintf("Description: %s\n\n", a.customQuestion.testDesc)
	
	s += "Questions:\n\n"
	for i, q := range a.customQuestion.questions {
		s += fmt.Sprintf("%d. %s\n", i+1, q.Text)
		s += fmt.Sprintf("   Type: %s\n", a.getQuestionTypeDisplay(q.Type))
		if len(q.Options) > 0 {
			s += "   Options: "
			for j, opt := range q.Options {
				if opt != "" {
					s += fmt.Sprintf("%c) %s ", 'A'+j, opt)
				}
			}
			s += "\n"
		}
		s += fmt.Sprintf("   Answer: %s\n", q.CorrectAnswer)
		if q.Explanation != "" {
			s += fmt.Sprintf("   Explanation: %s\n", q.Explanation)
		}
		s += "\n"
	}
	
	s += "Press Enter to save test to database\n"
	s += "Press 'b' to go back and add more questions\n"
	
	return s
}

// viewCustomQuestionInputMode renders input mode
func (a *App) viewCustomQuestionInputMode() string {
	var prompt string
	switch a.customQuestion.inputMode {
	case "test_name":
		prompt = "Enter test name:"
	case "test_desc":
		prompt = "Enter test description:"
	case "question":
		prompt = "Enter question text:"
	case "answer":
		prompt = "Enter correct answer:"
	case "explanation":
		prompt = "Enter explanation (optional):"
	case "option":
		prompt = fmt.Sprintf("Enter option %c:", 'A'+a.customQuestion.optionIndex)
	}
	
	s := prompt + "\n"
	s += "> " + a.customQuestion.input + "\n\n"
	s += "Press Enter to confirm, Esc to cancel\n"
	
	return s
}

// handleTestInfoStep handles test info step input
func (a *App) handleTestInfoStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if a.customQuestion.cursor > 0 {
			a.customQuestion.cursor--
		}
	case "down", "j":
		if a.customQuestion.cursor < 1 {
			a.customQuestion.cursor++
		}
	case "n":
		if a.customQuestion.cursor == 0 {
			a.customQuestion.inputMode = "test_name"
			a.customQuestion.input = a.customQuestion.testName
		}
	case "d":
		if a.customQuestion.cursor == 1 {
			a.customQuestion.inputMode = "test_desc"
			a.customQuestion.input = a.customQuestion.testDesc
		}
	case "enter", " ":
		a.customQuestion.step = 1
		a.customQuestion.cursor = 0
	}
	return a, nil
}

// handleQuestionCreationStep handles question creation step input
func (a *App) handleQuestionCreationStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if a.customQuestion.cursor > 0 {
			a.customQuestion.cursor--
		}
	case "down", "j":
		maxCursor := 4
		if a.customQuestion.cursor < maxCursor {
			a.customQuestion.cursor++
		}
	case "t":
		if a.customQuestion.cursor == 0 {
			a.cycleQuestionType()
		}
	case "q":
		if a.customQuestion.cursor == 1 {
			a.customQuestion.inputMode = "question"
			a.customQuestion.input = a.customQuestion.currentQuestion.text
		}
	case "o":
		if a.customQuestion.cursor == 2 && a.customQuestion.currentQuestion.qType == "multiple_choice" {
			a.customQuestion.inputMode = "option"
			a.customQuestion.optionIndex = 0
			a.customQuestion.input = a.customQuestion.currentQuestion.options[0]
		}
	case "a":
		if a.customQuestion.cursor == 3 {
			a.customQuestion.inputMode = "answer"
			a.customQuestion.input = a.customQuestion.currentQuestion.correctAnswer
		}
	case "e":
		if a.customQuestion.cursor == 4 {
			a.customQuestion.inputMode = "explanation"
			a.customQuestion.input = a.customQuestion.currentQuestion.explanation
		}
	case "s":
		return a.saveCurrentQuestion()
	case "f":
		if len(a.customQuestion.questions) > 0 {
			a.customQuestion.step = 2
			a.customQuestion.cursor = 0
		} else {
			a.customQuestion.errorMsg = "Create at least one question before finishing"
		}
	}
	return a, nil
}

// handleReviewStep handles review step input
func (a *App) handleReviewStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		return a.saveCustomTest()
	case "b":
		a.customQuestion.step = 1
		a.customQuestion.cursor = 0
	}
	return a, nil
}

// handleCustomQuestionInput handles input mode
func (a *App) handleCustomQuestionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm input
		switch a.customQuestion.inputMode {
		case "test_name":
			if err := a.validateInput(a.customQuestion.input, 1); err == nil {
				a.customQuestion.testName = strings.TrimSpace(a.customQuestion.input)
			} else {
				a.customQuestion.errorMsg = err.Error()
			}
		case "test_desc":
			a.customQuestion.testDesc = strings.TrimSpace(a.customQuestion.input)
		case "question":
			if err := a.validateInput(a.customQuestion.input, 5); err == nil {
				a.customQuestion.currentQuestion.text = strings.TrimSpace(a.customQuestion.input)
			} else {
				a.customQuestion.errorMsg = err.Error()
			}
		case "answer":
			if err := a.validateInput(a.customQuestion.input, 1); err == nil {
				a.customQuestion.currentQuestion.correctAnswer = strings.TrimSpace(a.customQuestion.input)
			} else {
				a.customQuestion.errorMsg = err.Error()
			}
		case "explanation":
			a.customQuestion.currentQuestion.explanation = strings.TrimSpace(a.customQuestion.input)
		case "option":
			if err := a.validateInput(a.customQuestion.input, 1); err == nil {
				a.customQuestion.currentQuestion.options[a.customQuestion.optionIndex] = strings.TrimSpace(a.customQuestion.input)
				// Move to next option or finish
				if a.customQuestion.optionIndex < 3 {
					a.customQuestion.optionIndex++
					a.customQuestion.input = a.customQuestion.currentQuestion.options[a.customQuestion.optionIndex]
					return a, nil // Stay in input mode for next option
				}
			} else {
				a.customQuestion.errorMsg = err.Error()
			}
		}
		a.customQuestion.inputMode = ""
		a.customQuestion.input = ""
	case "esc":
		// Cancel input
		a.customQuestion.inputMode = ""
		a.customQuestion.input = ""
	case "backspace":
		// Remove last character
		if len(a.customQuestion.input) > 0 {
			a.customQuestion.input = a.customQuestion.input[:len(a.customQuestion.input)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			a.customQuestion.input += msg.String()
		}
	}
	return a, nil
}

// cycleQuestionType cycles through question types
func (a *App) cycleQuestionType() {
	a.customQuestion.typeIndex = (a.customQuestion.typeIndex + 1) % len(a.customQuestion.questionTypes)
	a.customQuestion.currentQuestion.qType = a.customQuestion.questionTypes[a.customQuestion.typeIndex]
	
	// Reset options based on type
	switch a.customQuestion.currentQuestion.qType {
	case "multiple_choice":
		a.customQuestion.currentQuestion.options = make([]string, 4)
	case "true_false":
		a.customQuestion.currentQuestion.options = []string{}
	case "short_answer":
		a.customQuestion.currentQuestion.options = []string{}
	}
}

// saveCurrentQuestion saves the current question to the list
func (a *App) saveCurrentQuestion() (tea.Model, tea.Cmd) {
	// Validate question
	if strings.TrimSpace(a.customQuestion.currentQuestion.text) == "" {
		a.customQuestion.errorMsg = "Question text is required"
		return a, nil
	}
	
	if strings.TrimSpace(a.customQuestion.currentQuestion.correctAnswer) == "" {
		a.customQuestion.errorMsg = "Correct answer is required"
		return a, nil
	}
	
	// Validate multiple choice options
	if a.customQuestion.currentQuestion.qType == "multiple_choice" {
		validOptions := 0
		for _, opt := range a.customQuestion.currentQuestion.options {
			if strings.TrimSpace(opt) != "" {
				validOptions++
			}
		}
		if validOptions < 2 {
			a.customQuestion.errorMsg = "Multiple choice questions need at least 2 options"
			return a, nil
		}
	}
	
	// Save question
	question := QuestionData{
		Text:          strings.TrimSpace(a.customQuestion.currentQuestion.text),
		Type:          a.customQuestion.currentQuestion.qType,
		Options:       make([]string, len(a.customQuestion.currentQuestion.options)),
		CorrectAnswer: strings.TrimSpace(a.customQuestion.currentQuestion.correctAnswer),
		Explanation:   strings.TrimSpace(a.customQuestion.currentQuestion.explanation),
	}
	
	copy(question.Options, a.customQuestion.currentQuestion.options)
	a.customQuestion.questions = append(a.customQuestion.questions, question)
	
	// Reset current question
	a.customQuestion.currentQuestion.text = ""
	a.customQuestion.currentQuestion.correctAnswer = ""
	a.customQuestion.currentQuestion.explanation = ""
	if a.customQuestion.currentQuestion.qType == "multiple_choice" {
		a.customQuestion.currentQuestion.options = make([]string, 4)
	} else {
		a.customQuestion.currentQuestion.options = []string{}
	}
	
	a.customQuestion.successMsg = fmt.Sprintf("Question saved! (%d total)", len(a.customQuestion.questions))
	a.customQuestion.cursor = 0
	
	return a, nil
}

// saveCustomTest saves the custom test to database
func (a *App) saveCustomTest() (tea.Model, tea.Cmd) {
	if len(a.customQuestion.questions) == 0 {
		a.customQuestion.errorMsg = "No questions to save"
		return a, nil
	}
	
	// Create test in database
	test, err := a.db.CreateTest(a.customQuestion.testName, a.customQuestion.testDesc)
	if err != nil {
		a.customQuestion.errorMsg = fmt.Sprintf("Failed to create test: %v", err)
		return a, nil
	}
	
	// Save questions to database
	for _, q := range a.customQuestion.questions {
		_, err := a.db.CreateQuestion(test.ID, q.Text, q.Type, q.CorrectAnswer, q.Explanation, q.Options)
		if err != nil {
			a.customQuestion.errorMsg = fmt.Sprintf("Failed to save question: %v", err)
			return a, nil
		}
	}
	
	// Reset and return to main menu
	a.customQuestion = NewCustomQuestionModel()
	a.currentView = MainMenuView
	
	return a, nil
}