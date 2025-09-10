package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pdf-test-generator/chatgpt"
	"pdf-test-generator/database"
	"pdf-test-generator/pdf"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType represents different views in the application
type ViewType string

const (
	MainMenuView        ViewType = "main_menu"
	PDFProcessView      ViewType = "pdf_process"
	CustomQuestionView  ViewType = "custom_question"
	TestSelectionView   ViewType = "test_selection"
	TestTakingView      ViewType = "test_taking"
	TestResultsView     ViewType = "test_results"
	FileSelectionView   ViewType = "file_selection"
	QuestionGenView     ViewType = "question_gen"
)

// App represents the main application state
type App struct {
	currentView ViewType
	db          *database.DB
	chatGPT     *chatgpt.Client
	pdfProcessor *pdf.PDFProcessor
	
	// View models
	mainMenu        *MainMenuModel
	pdfProcess      *PDFProcessModel
	customQuestion  *CustomQuestionModel
	testSelection   *TestSelectionModel
	testTaking      *TestTakingModel
	testResults     *TestResultsModel
	fileSelection   *FileSelectionModel
	questionGen     *QuestionGenModel
	
	// Shared state
	currentTest     *database.Test
	currentQuestions []*database.Question
	userAnswers     map[int]string
	testStartTime   time.Time
}

// NewApp creates a new application instance
func NewApp(dbPath, apiKey string) (*App, error) {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app := &App{
		currentView:  MainMenuView,
		db:          db,
		chatGPT:     chatgpt.NewClient(apiKey),
		pdfProcessor: pdf.NewPDFProcessor(),
		userAnswers: make(map[int]string),
	}

	// Initialize view models
	app.mainMenu = NewMainMenuModel()
	app.pdfProcess = NewPDFProcessModel()
	app.customQuestion = NewCustomQuestionModel()
	app.testSelection = NewTestSelectionModel()
	app.testTaking = NewTestTakingModel()
	app.testResults = NewTestResultsModel()
	app.fileSelection = NewFileSelectionModel()
	app.questionGen = NewQuestionGenModel()

	return app, nil
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "esc":
			// Go back to main menu from any view
			if a.currentView != MainMenuView {
				a.currentView = MainMenuView
				return a, nil
			}
		}
	}

	// Route to appropriate view handler
	switch a.currentView {
	case MainMenuView:
		return a.updateMainMenu(msg)
	case PDFProcessView:
		return a.updatePDFProcess(msg)
	case CustomQuestionView:
		return a.updateCustomQuestion(msg)
	case TestSelectionView:
		return a.updateTestSelection(msg)
	case TestTakingView:
		return a.updateTestTaking(msg)
	case TestResultsView:
		return a.updateTestResults(msg)
	case FileSelectionView:
		return a.updateFileSelection(msg)
	case QuestionGenView:
		return a.updateQuestionGen(msg)
	default:
		return a, nil
	}
}

// View renders the current view
func (a *App) View() string {
	switch a.currentView {
	case MainMenuView:
		return a.viewMainMenu()
	case PDFProcessView:
		return a.viewPDFProcess()
	case CustomQuestionView:
		return a.viewCustomQuestion()
	case TestSelectionView:
		return a.viewTestSelection()
	case TestTakingView:
		return a.viewTestTaking()
	case TestResultsView:
		return a.viewTestResults()
	case FileSelectionView:
		return a.viewFileSelection()
	case QuestionGenView:
		return a.viewQuestionGen()
	default:
		return "Unknown view"
	}
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0099FF"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

// Helper functions
func (a *App) renderHeader(title string) string {
	return headerStyle.Render("📚 "+title) + "\n\n"
}

func (a *App) renderFooter() string {
	return "\n" + infoStyle.Render("Press 'esc' to go back to main menu, 'ctrl+c' to quit")
}

func (a *App) renderError(err string) string {
	return errorStyle.Render("❌ Error: "+err) + "\n"
}

func (a *App) renderSuccess(msg string) string {
	return successStyle.Render("✅ "+msg) + "\n"
}

// Navigation helpers
func (a *App) switchToView(view ViewType) tea.Cmd {
	a.currentView = view
	return nil
}

// File helper functions
func (a *App) listPDFFiles(dir string) ([]string, error) {
	var pdfFiles []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".pdf" {
			pdfFiles = append(pdfFiles, path)
		}
		
		return nil
	})
	
	return pdfFiles, err
}

// Question type helpers
func (a *App) getQuestionTypeDisplay(qType string) string {
	switch qType {
	case "multiple_choice":
		return "Multiple Choice"
	case "true_false":
		return "True/False"
	case "short_answer":
		return "Short Answer"
	default:
		return "Unknown"
	}
}

// Score calculation
func (a *App) calculateScore(questions []*database.Question, answers map[int]string) (int, float64) {
	correct := 0
	total := len(questions)
	
	for _, q := range questions {
		userAnswer, exists := answers[q.ID]
		if !exists {
			continue
		}
		
		// Normalize answers for comparison
		correctAnswer := strings.ToLower(strings.TrimSpace(q.CorrectAnswer))
		userAnswer = strings.ToLower(strings.TrimSpace(userAnswer))
		
		if correctAnswer == userAnswer {
			correct++
		}
	}
	
	score := 0.0
	if total > 0 {
		score = float64(correct) / float64(total) * 100
	}
	
	return correct, score
}

// Time formatting
func (a *App) formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// Input validation
func (a *App) validateInput(input string, minLength int) error {
	input = strings.TrimSpace(input)
	if len(input) < minLength {
		return fmt.Errorf("input must be at least %d characters long", minLength)
	}
	return nil
}

// Number parsing helper
func (a *App) parsePositiveInt(s string, defaultVal int) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && val > 0 {
		return val
	}
	return defaultVal
}