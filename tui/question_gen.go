package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// QuestionGenModel represents the question generation state
type QuestionGenModel struct {
	cursor      int
	status      string // "idle", "generating", "completed", "error"
	errorMsg    string
	successMsg  string
	progress    string
	generatedQuestions int
	totalQuestions     int
}

// NewQuestionGenModel creates a new question generation model
func NewQuestionGenModel() *QuestionGenModel {
	return &QuestionGenModel{
		status: "idle",
	}
}

// updateQuestionGen handles question generation updates
func (a *App) updateQuestionGen(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			a.currentView = MainMenuView
		case "r":
			// Restart generation if failed
			if a.questionGen.status == "error" {
				a.questionGen.status = "idle"
				a.questionGen.errorMsg = ""
			}
		}
	}
	return a, nil
}

// viewQuestionGen renders the question generation view
func (a *App) viewQuestionGen() string {
	s := a.renderHeader("Generating Questions")
	
	if a.questionGen.errorMsg != "" {
		s += a.renderError(a.questionGen.errorMsg)
	}
	
	if a.questionGen.successMsg != "" {
		s += a.renderSuccess(a.questionGen.successMsg)
	}
	
	switch a.questionGen.status {
	case "idle":
		s += "Ready to generate questions...\n\n"
	case "generating":
		s += "Generating questions from PDF content...\n\n"
		if a.questionGen.progress != "" {
			s += a.questionGen.progress + "\n\n"
		}
		if a.questionGen.totalQuestions > 0 {
			s += fmt.Sprintf("Progress: %d/%d questions generated\n\n", 
				a.questionGen.generatedQuestions, a.questionGen.totalQuestions)
		}
	case "completed":
		s += "Question generation completed successfully!\n\n"
		s += fmt.Sprintf("Generated %d questions\n\n", a.questionGen.generatedQuestions)
	case "error":
		s += "Question generation failed.\n\n"
		s += "Press 'r' to retry\n\n"
	}
	
	s += "Press 'q' to return to main menu\n"
	
	return s + a.renderFooter()
}

// startQuestionGeneration starts the question generation process
func (a *App) startQuestionGeneration(text string, numQuestions int) {
	a.questionGen.status = "generating"
	a.questionGen.generatedQuestions = 0
	a.questionGen.totalQuestions = numQuestions
	a.questionGen.progress = "Preparing to generate questions..."
	a.questionGen.errorMsg = ""
	a.questionGen.successMsg = ""
}

// updateGenerationProgress updates the generation progress
func (a *App) updateGenerationProgress(generated, total int, message string) {
	a.questionGen.generatedQuestions = generated
	a.questionGen.totalQuestions = total
	a.questionGen.progress = message
}

// completeGeneration marks generation as completed
func (a *App) completeGeneration(generated int) {
	a.questionGen.status = "completed"
	a.questionGen.generatedQuestions = generated
	a.questionGen.successMsg = fmt.Sprintf("Successfully generated %d questions!", generated)
}

// failGeneration marks generation as failed
func (a *App) failGeneration(err error) {
	a.questionGen.status = "error"
	a.questionGen.errorMsg = fmt.Sprintf("Generation failed: %v", err)
}