package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// FileSelectionModel represents the file selection state
type FileSelectionModel struct {
	files       []string
	cursor      int
	currentDir  string
	purpose     string // "pdf_generation" or other purposes
	errorMsg    string
	loading     bool
	inputMode   bool
	input       string
}

// NewFileSelectionModel creates a new file selection model
func NewFileSelectionModel() *FileSelectionModel {
	homeDir, _ := os.UserHomeDir()
	return &FileSelectionModel{
		currentDir: homeDir,
		files:      []string{},
	}
}

// updateFileSelection handles file selection updates
func (a *App) updateFileSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.fileSelection.inputMode {
			return a.handleFileInputMode(msg)
		}
		
		switch msg.String() {
		case "up", "k":
			if a.fileSelection.cursor > 0 {
				a.fileSelection.cursor--
			}
		case "down", "j":
			if a.fileSelection.cursor < len(a.fileSelection.files)-1 {
				a.fileSelection.cursor++
			}
		case "enter", " ":
			return a.handleFileSelection()
		case "r":
			// Refresh file list
			a.refreshFileList()
		case "c":
			// Change directory
			a.fileSelection.inputMode = true
			a.fileSelection.input = a.fileSelection.currentDir
		}
	}
	return a, nil
}

// viewFileSelection renders the file selection view
func (a *App) viewFileSelection() string {
	s := a.renderHeader("Select PDF File")
	
	if a.fileSelection.inputMode {
		s += "Enter directory path:\n"
		s += "> " + a.fileSelection.input + "\n\n"
		s += "Press Enter to confirm, Esc to cancel\n"
		return s + a.renderFooter()
	}
	
	if a.fileSelection.errorMsg != "" {
		s += a.renderError(a.fileSelection.errorMsg)
		a.fileSelection.errorMsg = ""
	}
	
	s += fmt.Sprintf("Current directory: %s\n\n", a.fileSelection.currentDir)
	
	if len(a.fileSelection.files) == 0 {
		s += "No PDF files found in this directory.\n\n"
		s += "Press 'c' to change directory, 'r' to refresh\n"
	} else {
		s += "PDF Files:\n\n"
		for i, file := range a.fileSelection.files {
			cursor := " "
			if a.fileSelection.cursor == i {
				cursor = ">"
				style := selectedStyle
				s += fmt.Sprintf("%s %s\n", cursor, style.Render(filepath.Base(file)))
			} else {
				s += fmt.Sprintf("%s %s\n", cursor, filepath.Base(file))
			}
		}
		s += "\nPress Enter to select, 'c' to change directory, 'r' to refresh\n"
	}
	
	return s + a.renderFooter()
}

// handleFileInputMode handles input mode for directory changes
func (a *App) handleFileInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm directory change
		if _, err := os.Stat(a.fileSelection.input); err == nil {
			a.fileSelection.currentDir = a.fileSelection.input
			a.refreshFileList()
		} else {
			a.fileSelection.errorMsg = "Directory does not exist"
		}
		a.fileSelection.inputMode = false
		a.fileSelection.input = ""
	case "esc":
		// Cancel directory change
		a.fileSelection.inputMode = false
		a.fileSelection.input = ""
	case "backspace":
		// Remove last character
		if len(a.fileSelection.input) > 0 {
			a.fileSelection.input = a.fileSelection.input[:len(a.fileSelection.input)-1]
		}
	default:
		// Add character to input
		if len(msg.String()) == 1 {
			a.fileSelection.input += msg.String()
		}
	}
	return a, nil
}

// handleFileSelection processes file selection
func (a *App) handleFileSelection() (tea.Model, tea.Cmd) {
	if len(a.fileSelection.files) == 0 {
		return a, nil
	}
	
	selectedFile := a.fileSelection.files[a.fileSelection.cursor]
	
	switch a.fileSelection.purpose {
	case "pdf_generation":
		// Process PDF for question generation
		a.pdfProcess.selectedFile = selectedFile
		a.currentView = PDFProcessView
		return a, nil
	default:
		return a, nil
	}
}

// refreshFileList refreshes the list of PDF files in current directory
func (a *App) refreshFileList() {
	files, err := a.listPDFFiles(a.fileSelection.currentDir)
	if err != nil {
		a.fileSelection.errorMsg = fmt.Sprintf("Error reading directory: %v", err)
		a.fileSelection.files = []string{}
	} else {
		a.fileSelection.files = files
	}
	a.fileSelection.cursor = 0
}

// Initialize file list when entering this view
func (a *App) initFileSelection() {
	if len(a.fileSelection.files) == 0 {
		a.refreshFileList()
	}
}