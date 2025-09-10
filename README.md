# PDF Test Generator

A Go-based terminal user interface (TUI) application that generates practice tests from PDF documents using ChatGPT and allows you to create custom questions for studying.

## Features

- **PDF Processing**: Extract text content from PDF files
- **AI Question Generation**: Generate questions automatically using ChatGPT API
- **Custom Question Creation**: Create your own questions with multiple choice, true/false, and short answer formats
- **Interactive Testing**: Take practice tests with a beautiful terminal interface
- **Results Tracking**: View detailed test results and performance analytics
- **SQLite Database**: Persistent storage for tests, questions, and results

## Prerequisites

- Go 1.19 or higher
- OpenAI API key (optional, for ChatGPT features)

## Installation

1. Clone or download this repository:
   ```bash
   git clone <repository-url>
   cd goStudyTest
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o pdf-test-generator
   ```

## Configuration

### Setting up API Keys

The application uses a `.env` file for configuration. This is the recommended way to set your API keys:

1. **Copy the example configuration file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit the `.env` file and add your API keys:**
   ```bash
   # Open the .env file in your preferred editor
   nano .env
   # or
   code .env
   ```

3. **Replace the placeholder with your actual OpenAI API key:**
   ```
   OPENAI_API_KEY=sk-your-actual-api-key-here
   ```

### Getting an OpenAI API Key

To use ChatGPT features for automatic question generation:

1. Visit [OpenAI's API Keys page](https://platform.openai.com/api-keys)
2. Sign in or create an account
3. Click "Create new secret key"
4. Copy the key and paste it in your `.env` file

### Alternative: System Environment Variables

You can also set environment variables directly (though `.env` file is recommended):

```bash
export OPENAI_API_KEY="your-api-key-here"
```

**Note**: The application will work without an API key, but ChatGPT features will be disabled. The `.env` file is ignored by git to keep your API keys secure.

## Usage

### Running the Application

```bash
./pdf-test-generator
```

Or run directly with Go:

```bash
go run main.go
```

### Main Menu Options

1. **📄 Generate questions from PDF**
   - Select a PDF file from your system
   - Extract text content automatically
   - Generate questions using ChatGPT
   - Review and save generated questions

2. **✏️ Create custom questions**
   - Create tests manually
   - Add multiple choice, true/false, or short answer questions
   - Set correct answers and explanations
   - Save custom tests to database

3. **📝 Take practice test**
   - Select from available tests
   - Interactive quiz interface
   - Real-time scoring
   - Detailed explanations for answers

4. **📊 View test results**
   - Review past test performance
   - Detailed answer breakdowns
   - Performance analytics
   - Delete old results

### Navigation

- **Arrow Keys** or **j/k**: Navigate up/down
- **Enter** or **Space**: Select/confirm
- **Esc**: Go back to previous screen
- **q**: Quit application (from main menu)
- **Ctrl+C**: Force quit from anywhere

## File Structure

```
.
├── main.go                 # Application entry point
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── README.md               # This file
├── .env.example            # Environment configuration template
├── .env                    # Your API keys (create from .env.example)
├── .gitignore              # Git ignore rules
├── chatgpt/
│   └── client.go           # OpenAI ChatGPT API client
├── database/
│   └── database.go         # SQLite database operations
├── pdf/
│   └── processor.go        # PDF text extraction
└── tui/
    ├── models.go           # Main TUI application state
    ├── main_menu.go        # Main menu interface
    ├── file_selection.go   # PDF file selection
    ├── pdf_process.go      # PDF processing interface
    ├── question_gen.go     # Question generation interface
    ├── custom_question.go  # Custom question creation
    ├── test_selection.go   # Test selection interface
    ├── test_taking.go      # Interactive test taking
    └── test_results.go     # Results viewing interface
```

## Database

The application uses SQLite for data persistence. The database file (`test_generator.db`) is created automatically in the application directory and contains:

- **tests**: Test metadata (name, description, creation date)
- **questions**: Individual questions with answers and explanations
- **test_results**: Test attempt results and scores
- **question_answers**: Detailed answers for each question attempt

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [pdf](https://github.com/ledongthuc/pdf) - PDF text extraction

## Question Types

### Multiple Choice
- Up to 4 answer options (A, B, C, D)
- Single correct answer
- Optional explanations

### True/False
- Simple true or false questions
- Optional explanations

### Short Answer
- Free-text responses
- Exact match scoring
- Case-insensitive comparison

## Tips for Best Results

### PDF Processing
- Use PDFs with selectable text (not scanned images)
- Ensure PDFs are not password-protected
- Academic papers and textbooks work best
- Clean, well-formatted documents produce better questions

### ChatGPT Question Generation
- Provide clear, educational content
- Longer text passages generate more diverse questions
- Review generated questions before saving
- Edit questions if needed for accuracy

### Custom Questions
- Write clear, unambiguous questions
- Provide helpful explanations
- Test your questions before using them
- Use varied question types for comprehensive testing

## Troubleshooting

### Common Issues

**Application won't start**
- Ensure Go is properly installed
- Check that all dependencies are downloaded: `go mod download`
- Verify the binary was built successfully: `go build`

**PDF processing fails**
- Ensure the PDF file is not corrupted
- Check file permissions
- Try with a different PDF file
- Verify the PDF contains selectable text

**ChatGPT features not working**
- Verify your OpenAI API key is set correctly
- Check your API key has sufficient credits
- Ensure you have internet connectivity
- Try with a smaller text sample first

**Database errors**
- Check write permissions in the application directory
- Delete `test_generator.db` to reset the database
- Ensure SQLite is properly installed

### Getting Help

If you encounter issues:

1. Check the error messages in the application
2. Verify your environment setup
3. Try with sample data first
4. Check the troubleshooting section above

## Development

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd goStudyTest

# Install dependencies
go mod download

# Run tests (if available)
go test ./...

# Build the application
go build -o pdf-test-generator

# Run the application
./pdf-test-generator
```

### Project Structure

The application follows a modular architecture:

- **main.go**: Entry point and application initialization
- **database/**: Data persistence layer
- **pdf/**: PDF processing functionality
- **chatgpt/**: AI integration for question generation
- **tui/**: Terminal user interface components

## License

This project is provided as-is for educational purposes.

## Contributing

Contributions are welcome! Please feel free to submit issues and enhancement requests.

---

**Happy studying! 📚**