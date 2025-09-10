package tui

import (
	"fmt"
	"time"

	"pdf-test-generator/database"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSelectionModel represents the test selection state
type TestSelectionModel struct {
	tests    []*database.Test
	cursor   int
	purpose  string // "take_test" or "view_tests"
	errorMsg string
	loading  bool
}

// NewTestSelectionModel creates a new test selection model
func NewTestSelectionModel() *TestSelectionModel {
	return &TestSelectionModel{
		tests: []*database.Test{},
	}
}

// updateTestSelection handles test selection updates
func (a *App) updateTestSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.testSelection.loading {
		return a, nil
	}
	
	// Load tests if not already loaded
	if len(a.testSelection.tests) == 0 {
		a.loadTests()
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if a.testSelection.cursor > 0 {
				a.testSelection.cursor--
			}
		case "down", "j":
			if a.testSelection.cursor < len(a.testSelection.tests)-1 {
				a.testSelection.cursor++
			}
		case "enter", " ":
			return a.handleTestSelection()
		case "d":
			// Delete selected test
			if len(a.testSelection.tests) > 0 {
				return a.deleteSelectedTest()
			}
		case "r":
			// Refresh test list
			a.loadTests()
		}
	}
	return a, nil
}

// viewTestSelection renders the test selection view
func (a *App) viewTestSelection() string {
	title := "Select Test"
	if a.testSelection.purpose == "view_tests" {
		title = "View Tests"
	}
	
	s := a.renderHeader(title)
	
	if a.testSelection.errorMsg != "" {
		s += a.renderError(a.testSelection.errorMsg)
		a.testSelection.errorMsg = ""
	}
	
	if a.testSelection.loading {
		s += "⏳ Loading tests...\n\n"
		return s + a.renderFooter()
	}
	
	if len(a.testSelection.tests) == 0 {
		s += "No tests found. Create some tests first!\n\n"
		s += "Press 'r' to refresh\n"
		return s + a.renderFooter()
	}
	
	s += "Available Tests:\n\n"
	
	for i, test := range a.testSelection.tests {
		cursor := " "
		if a.testSelection.cursor == i {
			cursor = ">"
			style := selectedStyle
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(a.formatTestInfo(test)))
		} else {
			s += fmt.Sprintf("%s %s\n", cursor, a.formatTestInfo(test))
		}
	}
	
	actionText := "take"
	if a.testSelection.purpose == "view_tests" {
		actionText = "view details for"
	}
	
	s += fmt.Sprintf("\nPress Enter to %s selected test, 'd' to delete, 'r' to refresh\n", actionText)
	
	return s + a.renderFooter()
}

// formatTestInfo formats test information for display
func (a *App) formatTestInfo(test *database.Test) string {
	// Get question count
	questions, _ := a.db.GetQuestionsByTestID(test.ID)
	questionCount := len(questions)
	
	// Format creation date
	createdDate := test.CreatedAt.Format("2006-01-02")
	
	return fmt.Sprintf("%s (%d questions) - Created: %s", test.Name, questionCount, createdDate)
}

// handleTestSelection processes test selection
func (a *App) handleTestSelection() (tea.Model, tea.Cmd) {
	if len(a.testSelection.tests) == 0 {
		return a, nil
	}
	
	selectedTest := a.testSelection.tests[a.testSelection.cursor]
	a.currentTest = selectedTest
	
	switch a.testSelection.purpose {
	case "take_test":
		// Load questions and start test
		questions, err := a.db.GetQuestionsByTestID(selectedTest.ID)
		if err != nil {
			a.testSelection.errorMsg = fmt.Sprintf("Failed to load questions: %v", err)
			return a, nil
		}
		
		if len(questions) == 0 {
			a.testSelection.errorMsg = "This test has no questions"
			return a, nil
		}
		
		a.currentQuestions = questions
		a.userAnswers = make(map[int]string)
		a.testStartTime = time.Now()
		a.testTaking.currentQuestion = 0
		a.testTaking.input = ""
		a.currentView = TestTakingView
		return a, nil
		
	case "view_tests":
		// Show test results/details
		a.currentView = TestResultsView
		a.loadTestResults()
		return a, nil
		
	default:
		return a, nil
	}
}

// loadTests loads all tests from database
func (a *App) loadTests() {
	a.testSelection.loading = true
	
	tests, err := a.db.GetAllTests()
	if err != nil {
		a.testSelection.errorMsg = fmt.Sprintf("Failed to load tests: %v", err)
		a.testSelection.tests = []*database.Test{}
	} else {
		a.testSelection.tests = tests
	}
	
	a.testSelection.cursor = 0
	a.testSelection.loading = false
}

// deleteSelectedTest deletes the currently selected test
func (a *App) deleteSelectedTest() (tea.Model, tea.Cmd) {
	if len(a.testSelection.tests) == 0 {
		return a, nil
	}
	
	selectedTest := a.testSelection.tests[a.testSelection.cursor]
	
	// Delete the test from database
	if err := a.db.DeleteTest(selectedTest.ID); err != nil {
		a.testSelection.errorMsg = fmt.Sprintf("Failed to delete test: %v", err)
		return a, nil
	}
	
	// Remove from local list
	a.testSelection.tests = append(a.testSelection.tests[:a.testSelection.cursor], a.testSelection.tests[a.testSelection.cursor+1:]...)
	
	// Adjust cursor if necessary
	if a.testSelection.cursor >= len(a.testSelection.tests) && len(a.testSelection.tests) > 0 {
		a.testSelection.cursor = len(a.testSelection.tests) - 1
	}
	
	return a, nil
}