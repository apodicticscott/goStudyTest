package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestResultsModel represents the test results view state
type TestResultsModel struct {
	cursor      int
	results     []TestResultData
	selectedResult *TestResultData
	viewMode    string // "list", "detail"
	errorMsg    string
	successMsg  string
}

// TestResultData represents a test result with details
type TestResultData struct {
	ID          int
	TestName    string
	Score       int
	TotalQuestions int
	Percentage  float64
	TimeTaken   time.Duration
	CompletedAt time.Time
	Answers     []AnswerData
}

// AnswerData represents an individual answer
type AnswerData struct {
	QuestionText  string
	UserAnswer    string
	CorrectAnswer string
	IsCorrect     bool
	Explanation   string
}

// NewTestResultsModel creates a new test results model
func NewTestResultsModel() *TestResultsModel {
	return &TestResultsModel{
		viewMode: "list",
	}
}

// updateTestResults handles test results updates
func (a *App) updateTestResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch a.testResults.viewMode {
		case "list":
			return a.handleResultsListInput(msg)
		case "detail":
			return a.handleResultsDetailInput(msg)
		}
	}
	return a, nil
}

// viewTestResults renders the test results view
func (a *App) viewTestResults() string {
	s := a.renderHeader("Test Results")
	
	if a.testResults.errorMsg != "" {
		s += a.renderError(a.testResults.errorMsg)
		a.testResults.errorMsg = ""
	}
	
	if a.testResults.successMsg != "" {
		s += a.renderSuccess(a.testResults.successMsg)
		a.testResults.successMsg = ""
	}
	
	switch a.testResults.viewMode {
	case "list":
		return s + a.viewResultsList() + a.renderFooter()
	case "detail":
		return s + a.viewResultsDetail() + a.renderFooter()
	default:
		return s + "Unknown view mode" + a.renderFooter()
	}
}

// viewResultsList renders the results list
func (a *App) viewResultsList() string {
	// Load results if not already loaded
	if a.testResults.results == nil {
		a.loadTestResults()
	}
	
	if len(a.testResults.results) == 0 {
		s := "No test results found.\n\n"
		s += "Take some practice tests to see your results here!\n\n"
		s += "Press 'q' to go back to main menu\n"
		return s
	}
	
	s := fmt.Sprintf("Found %d test result(s):\n\n", len(a.testResults.results))
	
	// Display results
	for i, result := range a.testResults.results {
		cursor := " "
		if i == a.testResults.cursor {
			cursor = ">"
		}
		
		percentage := float64(result.Score) / float64(result.TotalQuestions) * 100
		grade := a.getGrade(percentage)
		
		s += fmt.Sprintf("%s %s\n", cursor, result.TestName)
		s += fmt.Sprintf("   Score: %d/%d (%.1f%%) - %s\n", 
			result.Score, result.TotalQuestions, percentage, grade)
		s += fmt.Sprintf("   Completed: %s\n", 
			result.CompletedAt.Format("Jan 2, 2006 3:04 PM"))
		if result.TimeTaken > 0 {
			s += fmt.Sprintf("   Time: %s\n", a.formatDuration(result.TimeTaken))
		}
		s += "\n"
	}
	
	s += "Press Enter to view detailed results\n"
	s += "Press 'd' to delete selected result\n"
	s += "Press 'r' to refresh results\n"
	s += "Use arrow keys to navigate\n"
	
	return s
}

// viewResultsDetail renders detailed results
func (a *App) viewResultsDetail() string {
	if a.testResults.selectedResult == nil {
		return "No result selected\n"
	}
	
	result := a.testResults.selectedResult
	percentage := float64(result.Score) / float64(result.TotalQuestions) * 100
	grade := a.getGrade(percentage)
	
	s := fmt.Sprintf("Test: %s\n", result.TestName)
	s += fmt.Sprintf("Score: %d/%d (%.1f%%) - %s\n", 
		result.Score, result.TotalQuestions, percentage, grade)
	s += fmt.Sprintf("Completed: %s\n", 
		result.CompletedAt.Format("Jan 2, 2006 3:04 PM"))
	if result.TimeTaken > 0 {
		s += fmt.Sprintf("Time Taken: %s\n", a.formatDuration(result.TimeTaken))
	}
	s += "\n"
	
	if len(result.Answers) == 0 {
		s += "No detailed answers available.\n"
	} else {
		s += "Question Details:\n\n"
		
		for i, answer := range result.Answers {
			status := "✗"
			if answer.IsCorrect {
				status = "✓"
			}
			
			s += fmt.Sprintf("%d. %s %s\n", i+1, status, answer.QuestionText)
			s += fmt.Sprintf("   Your Answer: %s\n", answer.UserAnswer)
			if !answer.IsCorrect {
				s += fmt.Sprintf("   Correct Answer: %s\n", answer.CorrectAnswer)
			}
			if answer.Explanation != "" {
				s += fmt.Sprintf("   Explanation: %s\n", answer.Explanation)
			}
			s += "\n"
		}
	}
	
	s += "Press 'b' to go back to results list\n"
	s += "Press 'd' to delete this result\n"
	
	return s
}

// handleResultsListInput handles input in list mode
func (a *App) handleResultsListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if a.testResults.cursor > 0 {
			a.testResults.cursor--
		}
	case "down", "j":
		if a.testResults.cursor < len(a.testResults.results)-1 {
			a.testResults.cursor++
		}
	case "enter", " ":
		if len(a.testResults.results) > 0 {
			a.testResults.selectedResult = &a.testResults.results[a.testResults.cursor]
			a.loadResultDetails(a.testResults.selectedResult)
			a.testResults.viewMode = "detail"
		}
	case "d":
		if len(a.testResults.results) > 0 {
			return a.deleteTestResult()
		}
	case "r":
		a.loadTestResults()
		a.testResults.successMsg = "Results refreshed"
	case "q":
		a.currentView = MainMenuView
	}
	return a, nil
}

// handleResultsDetailInput handles input in detail mode
func (a *App) handleResultsDetailInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		a.testResults.viewMode = "list"
		a.testResults.selectedResult = nil
	case "d":
		return a.deleteTestResult()
	case "q":
		a.currentView = MainMenuView
	}
	return a, nil
}

// loadTestResults loads test results from database
func (a *App) loadTestResults() {
	results, err := a.db.GetAllTestResults()
	if err != nil {
		a.testResults.errorMsg = fmt.Sprintf("Failed to load results: %v", err)
		return
	}
	
	// Convert database results to display format
	a.testResults.results = make([]TestResultData, len(results))
	for i, result := range results {
		a.testResults.results[i] = TestResultData{
			ID:             result.ID,
			TestName:       result.TestName,
			Score:          int(result.Score),
			TotalQuestions: result.TotalQuestions,
			Percentage:     result.Score / float64(result.TotalQuestions) * 100,
			TimeTaken:      time.Duration(result.TimeTaken) * time.Second,
			CompletedAt:    result.CompletedAt,
		}
	}
	
	// Reset cursor if out of bounds
	if a.testResults.cursor >= len(a.testResults.results) {
		a.testResults.cursor = 0
	}
}

// loadResultDetails loads detailed answers for a result
func (a *App) loadResultDetails(result *TestResultData) {
	answers, err := a.db.GetTestResultAnswers(result.ID)
	if err != nil {
		a.testResults.errorMsg = fmt.Sprintf("Failed to load result details: %v", err)
		return
	}
	
	// Convert database answers to display format
	result.Answers = make([]AnswerData, len(answers))
	for i, answer := range answers {
		result.Answers[i] = AnswerData{
			QuestionText:  answer.QuestionText,
			UserAnswer:    answer.UserAnswer,
			CorrectAnswer: answer.CorrectAnswer,
			IsCorrect:     answer.IsCorrect,
			Explanation:   answer.Explanation,
		}
	}
}

// deleteTestResult deletes the selected test result
func (a *App) deleteTestResult() (tea.Model, tea.Cmd) {
	var resultID int
	var testName string
	
	if a.testResults.viewMode == "detail" && a.testResults.selectedResult != nil {
		resultID = a.testResults.selectedResult.ID
		testName = a.testResults.selectedResult.TestName
	} else if a.testResults.viewMode == "list" && len(a.testResults.results) > 0 {
		resultID = a.testResults.results[a.testResults.cursor].ID
		testName = a.testResults.results[a.testResults.cursor].TestName
	} else {
		a.testResults.errorMsg = "No result selected for deletion"
		return a, nil
	}
	
	err := a.db.DeleteTestResult(resultID)
	if err != nil {
		a.testResults.errorMsg = fmt.Sprintf("Failed to delete result: %v", err)
		return a, nil
	}
	
	// Refresh results and return to list view
	a.loadTestResults()
	a.testResults.viewMode = "list"
	a.testResults.selectedResult = nil
	a.testResults.successMsg = fmt.Sprintf("Deleted result for '%s'", testName)
	
	return a, nil
}

// getGrade returns a letter grade based on percentage
func (a *App) getGrade(percentage float64) string {
	switch {
	case percentage >= 90:
		return "A"
	case percentage >= 80:
		return "B"
	case percentage >= 70:
		return "C"
	case percentage >= 60:
		return "D"
	default:
		return "F"
	}
}