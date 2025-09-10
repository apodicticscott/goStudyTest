package tui

import (
	"fmt"
	"strings"
	"time"

	"pdf-test-generator/database"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTakingModel represents the test taking state
type TestTakingModel struct {
	currentQuestion int
	input           string
	inputMode       bool
	showResult      bool
	resultMsg       string
	cursor          int // For multiple choice options
	errorMsg        string
	// Answer review functionality
	reviewMode     bool
	reviewQuestion int
}

// NewTestTakingModel creates a new test taking model
func NewTestTakingModel() *TestTakingModel {
	return &TestTakingModel{}
}

// updateTestTaking handles test taking updates
func (a *App) updateTestTaking(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(a.currentQuestions) == 0 {
		a.currentView = MainMenuView
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.testTaking.showResult {
			return a.handleResultView(msg)
		}

		currentQ := a.currentQuestions[a.testTaking.currentQuestion]

		switch currentQ.QuestionType {
		case "multiple_choice":
			return a.handleMultipleChoice(msg)
		case "true_false":
			return a.handleTrueFalse(msg)
		case "short_answer":
			return a.handleShortAnswer(msg)
		}
	}
	return a, nil
}

// viewTestTaking renders the test taking view
func (a *App) viewTestTaking() string {
	if len(a.currentQuestions) == 0 {
		return "No questions available"
	}

	s := a.renderHeader(fmt.Sprintf("Taking Test: %s", a.currentTest.Name))

	if a.testTaking.errorMsg != "" {
		s += a.renderError(a.testTaking.errorMsg)
		a.testTaking.errorMsg = ""
	}

	if a.testTaking.showResult {
		return s + a.viewTestComplete() + a.renderFooter()
	}

	// Progress indicator
	progress := fmt.Sprintf("Question %d of %d", a.testTaking.currentQuestion+1, len(a.currentQuestions))
	elapsed := time.Since(a.testStartTime)
	s += fmt.Sprintf("%s | Time: %s\n\n", progress, a.formatDuration(elapsed))

	currentQ := a.currentQuestions[a.testTaking.currentQuestion]
	s += fmt.Sprintf("Q%d: %s\n\n", a.testTaking.currentQuestion+1, currentQ.QuestionText)

	switch currentQ.QuestionType {
	case "multiple_choice":
		s += a.viewMultipleChoice(currentQ)
	case "true_false":
		s += a.viewTrueFalse()
	case "short_answer":
		s += a.viewShortAnswer()
	}

	return s + a.renderFooter()
}

// viewMultipleChoice renders multiple choice question
func (a *App) viewMultipleChoice(question *database.Question) string {
	s := "Choose the correct answer:\n\n"

	letters := []string{"A", "B", "C", "D"}
	for i, option := range question.Options {
		if i >= len(letters) {
			break
		}

		cursor := "  "
		if a.testTaking.cursor == i {
			cursor = "► "
			style := selectedStyle
			s += fmt.Sprintf("%s%s) %s\n", cursor, letters[i], style.Render(option))
		} else {
			s += fmt.Sprintf("%s%s) %s\n", cursor, letters[i], option)
		}
	}

	s += "\n↑↓ Navigate • Enter/Space to select\n"
	return s
}

// viewTrueFalse renders true/false question
func (a *App) viewTrueFalse() string {
	s := "Select True or False:\n\n"

	options := []string{"True", "False"}
	for i, option := range options {
		cursor := "  "
		if a.testTaking.cursor == i {
			cursor = "► "
			style := selectedStyle
			s += fmt.Sprintf("%s%s\n", cursor, style.Render(option))
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, option)
		}
	}

	s += "\n↑↓ Navigate • Enter/Space to select\n"
	return s
}

// viewShortAnswer renders short answer question
func (a *App) viewShortAnswer() string {
	s := "Enter your answer:\n\n"
	s += "> " + a.testTaking.input + "\n\n"
	s += "Type your answer and press Enter to confirm\n"
	return s
}

// viewTestComplete renders test completion screen
func (a *App) viewTestComplete() string {
	if a.testTaking.reviewMode {
		return a.viewAnswerReview()
	}

	correct, score := a.calculateScore(a.currentQuestions, a.userAnswers)
	total := len(a.currentQuestions)
	elapsed := time.Since(a.testStartTime)

	s := "🎉 Test Complete! 🎉\n\n"
	s += fmt.Sprintf("Score: %.1f%% (%d/%d correct)\n", score, correct, total)
	s += fmt.Sprintf("Time taken: %s\n\n", a.formatDuration(elapsed))

	if a.testTaking.resultMsg != "" {
		s += a.testTaking.resultMsg + "\n\n"
	}

	s += "Press Enter to save results and return to main menu\n"
	s += "Press 'r' to review answers\n"

	return s
}

// handleMultipleChoice handles multiple choice input
func (a *App) handleMultipleChoice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentQ := a.currentQuestions[a.testTaking.currentQuestion]

	switch msg.String() {
	case "up", "k":
		if a.testTaking.cursor > 0 {
			a.testTaking.cursor--
		}
	case "down", "j":
		if a.testTaking.cursor < len(currentQ.Options)-1 {
			a.testTaking.cursor++
		}
	case "enter", " ":
		if len(currentQ.Options) > a.testTaking.cursor {
			// Store answer as letter (A, B, C, D)
			letters := []string{"A", "B", "C", "D"}
			if a.testTaking.cursor < len(letters) {
				a.userAnswers[currentQ.ID] = letters[a.testTaking.cursor]
				return a.nextQuestion()
			}
		}
	}
	return a, nil
}

// handleTrueFalse handles true/false input
func (a *App) handleTrueFalse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentQ := a.currentQuestions[a.testTaking.currentQuestion]

	switch msg.String() {
	case "up", "k":
		if a.testTaking.cursor > 0 {
			a.testTaking.cursor--
		}
	case "down", "j":
		if a.testTaking.cursor < 1 {
			a.testTaking.cursor++
		}
	case "enter", " ":
		answer := "true"
		if a.testTaking.cursor == 1 {
			answer = "false"
		}
		a.userAnswers[currentQ.ID] = answer
		return a.nextQuestion()
	}
	return a, nil
}

// handleShortAnswer handles short answer input
func (a *App) handleShortAnswer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentQ := a.currentQuestions[a.testTaking.currentQuestion]

	switch msg.String() {
	case "enter":
		if strings.TrimSpace(a.testTaking.input) == "" {
			a.testTaking.errorMsg = "Please enter an answer"
			return a, nil
		}
		a.userAnswers[currentQ.ID] = strings.TrimSpace(a.testTaking.input)
		a.testTaking.input = ""
		return a.nextQuestion()
	case "backspace":
		if len(a.testTaking.input) > 0 {
			a.testTaking.input = a.testTaking.input[:len(a.testTaking.input)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			a.testTaking.input += msg.String()
		}
	}
	return a, nil
}

// handleResultView handles input in result view
func (a *App) handleResultView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.testTaking.reviewMode {
		return a.handleAnswerReview(msg)
	}

	switch msg.String() {
	case "enter":
		// Save results and return to main menu
		return a.saveTestResults()
	case "r":
		// Start answer review
		a.testTaking.reviewMode = true
		a.testTaking.reviewQuestion = 0
	}
	return a, nil
}

// nextQuestion moves to the next question or completes the test
func (a *App) nextQuestion() (tea.Model, tea.Cmd) {
	a.testTaking.cursor = 0

	if a.testTaking.currentQuestion < len(a.currentQuestions)-1 {
		// Move to next question
		a.testTaking.currentQuestion++
	} else {
		// Test complete
		a.testTaking.showResult = true
	}

	return a, nil
}

// viewAnswerReview renders the answer review screen
func (a *App) viewAnswerReview() string {
	if len(a.currentQuestions) == 0 {
		return "No questions to review"
	}

	currentQ := a.currentQuestions[a.testTaking.reviewQuestion]
	userAnswer := a.userAnswers[currentQ.ID]
	correctAnswer := currentQ.CorrectAnswer
	isCorrect := strings.EqualFold(userAnswer, correctAnswer)

	s := a.renderHeader(fmt.Sprintf("Answer Review - Question %d of %d", a.testTaking.reviewQuestion+1, len(a.currentQuestions)))

	// Question
	s += fmt.Sprintf("Q%d: %s\n\n", a.testTaking.reviewQuestion+1, currentQ.QuestionText)

	// Show options for multiple choice
	if currentQ.QuestionType == "multiple_choice" {
		letters := []string{"A", "B", "C", "D"}
		for i, option := range currentQ.Options {
			if i >= len(letters) {
				break
			}

			prefix := fmt.Sprintf("  %s) ", letters[i])
			if letters[i] == userAnswer {
				if isCorrect {
					prefix = fmt.Sprintf("✓ %s) ", letters[i])
					s += successStyle.Render(prefix+option) + "\n"
				} else {
					prefix = fmt.Sprintf("✗ %s) ", letters[i])
					s += errorStyle.Render(prefix+option) + "\n"
				}
			} else if letters[i] == correctAnswer {
				prefix = fmt.Sprintf("✓ %s) ", letters[i])
				s += successStyle.Render(prefix+option) + "\n"
			} else {
				s += prefix + option + "\n"
			}
		}
	} else {
		// For true/false and short answer
		s += fmt.Sprintf("Your answer: %s\n", userAnswer)
		s += fmt.Sprintf("Correct answer: %s\n", correctAnswer)
	}

	s += "\n"

	// Result indicator
	if isCorrect {
		s += successStyle.Render("✓ CORRECT") + "\n\n"
	} else {
		s += errorStyle.Render("✗ INCORRECT") + "\n\n"
	}

	// Show explanation if available
	if currentQ.Explanation != "" {
		s += "Explanation:\n"
		s += infoStyle.Render(currentQ.Explanation) + "\n\n"
	}

	// Navigation instructions
	s += "← → Navigate questions • Esc to return to results\n"

	return s + a.renderFooter()
}

// handleAnswerReview handles input during answer review
func (a *App) handleAnswerReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if a.testTaking.reviewQuestion > 0 {
			a.testTaking.reviewQuestion--
		}
	case "right", "l":
		if a.testTaking.reviewQuestion < len(a.currentQuestions)-1 {
			a.testTaking.reviewQuestion++
		}
	case "esc":
		// Exit review mode
		a.testTaking.reviewMode = false
		a.testTaking.reviewQuestion = 0
	}
	return a, nil
}

// saveTestResults saves the test results to database
func (a *App) saveTestResults() (tea.Model, tea.Cmd) {
	correct, score := a.calculateScore(a.currentQuestions, a.userAnswers)
	total := len(a.currentQuestions)
	timeTaken := int(time.Since(a.testStartTime).Seconds())

	// Save test result
	result, err := a.db.SaveTestResult(a.currentTest.ID, score, total, correct, timeTaken)
	if err != nil {
		a.testTaking.errorMsg = fmt.Sprintf("Failed to save results: %v", err)
		return a, nil
	}

	// Save individual question answers (simplified - not implementing detailed answer tracking for now)
	_ = result // Use result if needed for detailed tracking

	// Reset state and return to main menu
	a.testTaking = NewTestTakingModel()
	a.currentTest = nil
	a.currentQuestions = nil
	a.userAnswers = make(map[int]string)
	a.currentView = MainMenuView

	return a, nil
}
