package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// MainMenuModel represents the main menu state
type MainMenuModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// NewMainMenuModel creates a new main menu model
func NewMainMenuModel() *MainMenuModel {
	return &MainMenuModel{
		choices: []string{
			"📄 Generate questions from PDF",
			"✏️  Create custom questions",
			"📝 Take practice test",
			"📊 View saved tests",
			"🚪 Exit",
		},
		selected: make(map[int]struct{}),
	}
}

// updateMainMenu handles main menu updates
func (a *App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return a, tea.Quit
		case "up", "k":
			if a.mainMenu.cursor > 0 {
				a.mainMenu.cursor--
			}
		case "down", "j":
			if a.mainMenu.cursor < len(a.mainMenu.choices)-1 {
				a.mainMenu.cursor++
			}
		case "enter", " ":
			return a.handleMainMenuSelection()
		}
	}
	return a, nil
}

// viewMainMenu renders the main menu
func (a *App) viewMainMenu() string {
	s := a.renderHeader("PDF Test Generator")
	s += "What would you like to do?\n\n"

	for i, choice := range a.mainMenu.choices {
		cursor := " "
		if a.mainMenu.cursor == i {
			cursor = ">"
			style := selectedStyle
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
		} else {
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	}

	s += "\nPress 'q' to quit, arrow keys to navigate, enter to select.\n"
	return s
}

// handleMainMenuSelection processes main menu selections
func (a *App) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	switch a.mainMenu.cursor {
	case 0:
		// Generate questions from PDF
		a.currentView = FileSelectionView
		a.fileSelection.purpose = "pdf_generation"
		return a, nil
	case 1:
		// Create custom questions
		a.currentView = CustomQuestionView
		return a, nil
	case 2:
		// Take practice test
		a.currentView = TestSelectionView
		a.testSelection.purpose = "take_test"
		return a, nil
	case 3:
		// View saved tests
		a.currentView = TestSelectionView
		a.testSelection.purpose = "view_tests"
		return a, nil
	case 4:
		// Exit
		return a, tea.Quit
	}
	return a, nil
}