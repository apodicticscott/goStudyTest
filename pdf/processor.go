package pdf

import (
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFProcessor handles PDF text extraction
type PDFProcessor struct{}

// NewPDFProcessor creates a new PDF processor
func NewPDFProcessor() *PDFProcessor {
	return &PDFProcessor{}
}

// ExtractText extracts text content from a PDF file
func (processor *PDFProcessor) ExtractText(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	var textBuilder strings.Builder
	totalPages := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := r.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		// Extract text from the page
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			// Continue with other pages if one fails
			continue
		}

		// Clean and format the text
		cleanedText := processor.cleanText(pageText)
		if cleanedText != "" {
			textBuilder.WriteString(cleanedText)
			textBuilder.WriteString("\n\n")
		}
	}

	extractedText := textBuilder.String()
	if extractedText == "" {
		return "", fmt.Errorf("no text could be extracted from the PDF")
	}

	return strings.TrimSpace(extractedText), nil
}

// cleanText cleans and formats extracted text
func (processor *PDFProcessor) cleanText(text string) string {
	// Remove excessive whitespace
	lines := strings.Split(text, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Trim whitespace from each line
		line = strings.TrimSpace(line)
		
		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip lines that are too short (likely artifacts)
		if len(line) < 3 {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, " ")
}

// GetTextSummary returns a summary of the extracted text for preview
func (processor *PDFProcessor) GetTextSummary(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// Find a good breaking point near the max length
	breakPoint := maxLength
	for i := maxLength; i > maxLength-50 && i > 0; i-- {
		if text[i] == ' ' || text[i] == '.' || text[i] == '\n' {
			breakPoint = i
			break
		}
	}

	return text[:breakPoint] + "..."
}

// ValidatePDF checks if a file is a valid PDF
func (processor *PDFProcessor) ValidatePDF(filePath string) error {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("invalid PDF file: %w", err)
	}
	defer f.Close()

	if r.NumPage() == 0 {
		return fmt.Errorf("PDF file has no pages")
	}

	return nil
}

// GetPDFInfo returns basic information about the PDF
func (processor *PDFProcessor) GetPDFInfo(filePath string) (map[string]interface{}, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	info := map[string]interface{}{
		"pages": r.NumPage(),
		"title": "",
	}

	// Try to get document info
	if r.Trailer().Key("Info").Kind() != pdf.Null {
		infoDict := r.Trailer().Key("Info")
		if !infoDict.IsNull() {
			if title := infoDict.Key("Title"); !title.IsNull() {
				info["title"] = title.Text()
			}
		}
	}

	return info, nil
}