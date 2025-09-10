package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	*sql.DB
}

// Test represents a practice test
type Test struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Question represents a test question
type Question struct {
	ID            int      `json:"id"`
	TestID        int      `json:"test_id"`
	QuestionText  string   `json:"question_text"`
	QuestionType  string   `json:"question_type"` // "multiple_choice", "true_false", "short_answer"
	Options       []string `json:"options"`        // For multiple choice questions
	CorrectAnswer string   `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
	CreatedAt     time.Time `json:"created_at"`
}

// TestResult represents a test attempt result
type TestResult struct {
	ID          int       `json:"id"`
	TestID      int       `json:"test_id"`
	Score       float64   `json:"score"`
	TotalQuestions int    `json:"total_questions"`
	CorrectAnswers int    `json:"correct_answers"`
	TimeTaken   int       `json:"time_taken"` // in seconds
	CompletedAt time.Time `json:"completed_at"`
}

// QuestionAnswer represents a user's answer to a question
type QuestionAnswer struct {
	ID           int    `json:"id"`
	ResultID     int    `json:"result_id"`
	QuestionID   int    `json:"question_id"`
	UserAnswer   string `json:"user_answer"`
	IsCorrect    bool   `json:"is_correct"`
}

// NewDB creates a new database connection and initializes tables
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbWrapper := &DB{db}
	if err := dbWrapper.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbWrapper, nil
}

// createTables creates the necessary database tables
func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS questions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			test_id INTEGER NOT NULL,
			question_text TEXT NOT NULL,
			question_type TEXT NOT NULL CHECK(question_type IN ('multiple_choice', 'true_false', 'short_answer')),
			options TEXT, -- JSON array for multiple choice options
			correct_answer TEXT NOT NULL,
			explanation TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS test_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			test_id INTEGER NOT NULL,
			score REAL NOT NULL,
			total_questions INTEGER NOT NULL,
			correct_answers INTEGER NOT NULL,
			time_taken INTEGER NOT NULL, -- in seconds
			completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (test_id) REFERENCES tests(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS question_answers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			result_id INTEGER NOT NULL,
			question_id INTEGER NOT NULL,
			user_answer TEXT NOT NULL,
			is_correct BOOLEAN NOT NULL,
			FOREIGN KEY (result_id) REFERENCES test_results(id) ON DELETE CASCADE,
			FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// CreateTest creates a new test
func (db *DB) CreateTest(name, description string) (*Test, error) {
	query := `INSERT INTO tests (name, description) VALUES (?, ?)`
	result, err := db.Exec(query, name, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create test: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return db.GetTest(int(id))
}

// GetTest retrieves a test by ID
func (db *DB) GetTest(id int) (*Test, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM tests WHERE id = ?`
	row := db.QueryRow(query, id)

	var test Test
	err := row.Scan(&test.ID, &test.Name, &test.Description, &test.CreatedAt, &test.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get test: %w", err)
	}

	return &test, nil
}

// GetAllTests retrieves all tests
func (db *DB) GetAllTests() ([]*Test, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM tests ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tests: %w", err)
	}
	defer rows.Close()

	var tests []*Test
	for rows.Next() {
		var test Test
		err := rows.Scan(&test.ID, &test.Name, &test.Description, &test.CreatedAt, &test.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test: %w", err)
		}
		tests = append(tests, &test)
	}

	return tests, nil
}

// CreateQuestion creates a new question for a test
func (db *DB) CreateQuestion(testID int, questionText, questionType, correctAnswer, explanation string, options []string) (*Question, error) {
	// Convert options to JSON string if provided
	var optionsJSON string
	if len(options) > 0 {
		// Simple JSON encoding for options
		optionsJSON = "[\"" + options[0]
		for i := 1; i < len(options); i++ {
			optionsJSON += "\",\"" + options[i]
		}
		optionsJSON += "\"]"
	}

	query := `INSERT INTO questions (test_id, question_text, question_type, options, correct_answer, explanation) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, testID, questionText, questionType, optionsJSON, correctAnswer, explanation)
	if err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return db.GetQuestion(int(id))
}

// GetQuestion retrieves a question by ID
func (db *DB) GetQuestion(id int) (*Question, error) {
	query := `SELECT id, test_id, question_text, question_type, options, correct_answer, explanation, created_at FROM questions WHERE id = ?`
	row := db.QueryRow(query, id)

	var question Question
	var optionsJSON string
	err := row.Scan(&question.ID, &question.TestID, &question.QuestionText, &question.QuestionType, &optionsJSON, &question.CorrectAnswer, &question.Explanation, &question.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	// Parse options JSON
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &question.Options); err != nil {
			// Fallback to empty options if JSON parsing fails
			question.Options = []string{}
		}
	}

	return &question, nil
}

// GetQuestionsByTestID retrieves all questions for a test
func (db *DB) GetQuestionsByTestID(testID int) ([]*Question, error) {
	query := `SELECT id, test_id, question_text, question_type, options, correct_answer, explanation, created_at FROM questions WHERE test_id = ? ORDER BY id`
	rows, err := db.Query(query, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	defer rows.Close()

	var questions []*Question
	for rows.Next() {
		var question Question
		var optionsJSON string
		err := rows.Scan(&question.ID, &question.TestID, &question.QuestionText, &question.QuestionType, &optionsJSON, &question.CorrectAnswer, &question.Explanation, &question.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		// Parse options JSON
		if optionsJSON != "" {
			if err := json.Unmarshal([]byte(optionsJSON), &question.Options); err != nil {
				// Fallback to empty options if JSON parsing fails
				question.Options = []string{}
			}
		}

		questions = append(questions, &question)
	}

	return questions, nil
}

// SaveTestResult saves a test result
func (db *DB) SaveTestResult(testID int, score float64, totalQuestions, correctAnswers, timeTaken int) (*TestResult, error) {
	query := `INSERT INTO test_results (test_id, score, total_questions, correct_answers, time_taken) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, testID, score, totalQuestions, correctAnswers, timeTaken)
	if err != nil {
		return nil, fmt.Errorf("failed to save test result: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &TestResult{
		ID:             int(id),
		TestID:         testID,
		Score:          score,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctAnswers,
		TimeTaken:      timeTaken,
		CompletedAt:    time.Now(),
	}, nil
}

// GetTestResults retrieves all results for a test
func (db *DB) GetTestResults(testID int) ([]*TestResult, error) {
	query := `SELECT id, test_id, score, total_questions, correct_answers, time_taken, completed_at FROM test_results WHERE test_id = ? ORDER BY completed_at DESC`
	rows, err := db.Query(query, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test results: %w", err)
	}
	defer rows.Close()

	var results []*TestResult
	for rows.Next() {
		var result TestResult
		err := rows.Scan(&result.ID, &result.TestID, &result.Score, &result.TotalQuestions, &result.CorrectAnswers, &result.TimeTaken, &result.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test result: %w", err)
		}
		results = append(results, &result)
	}

	return results, nil
}

// TestResultWithName represents a test result with test name
type TestResultWithName struct {
	ID             int       `json:"id"`
	TestID         int       `json:"test_id"`
	TestName       string    `json:"test_name"`
	Score          float64   `json:"score"`
	TotalQuestions int       `json:"total_questions"`
	CorrectAnswers int       `json:"correct_answers"`
	TimeTaken      int       `json:"time_taken"`
	CompletedAt    time.Time `json:"completed_at"`
}

// QuestionAnswerDetail represents a detailed question answer
type QuestionAnswerDetail struct {
	ID            int    `json:"id"`
	ResultID      int    `json:"result_id"`
	QuestionID    int    `json:"question_id"`
	QuestionText  string `json:"question_text"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	IsCorrect     bool   `json:"is_correct"`
	Explanation   string `json:"explanation"`
}

// GetAllTestResults returns all test results with test names
func (db *DB) GetAllTestResults() ([]*TestResultWithName, error) {
	rows, err := db.Query(`
		SELECT tr.id, tr.test_id, t.name, tr.score, tr.total_questions, tr.correct_answers, tr.time_taken, tr.completed_at
		FROM test_results tr
		JOIN tests t ON tr.test_id = t.id
		ORDER BY tr.completed_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all test results: %w", err)
	}
	defer rows.Close()

	var results []*TestResultWithName
	for rows.Next() {
		result := &TestResultWithName{}
		err := rows.Scan(&result.ID, &result.TestID, &result.TestName, &result.Score, &result.TotalQuestions, &result.CorrectAnswers, &result.TimeTaken, &result.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test result: %w", err)
		}
		results = append(results, result)
	}
	return results, nil
}

// GetTestResultAnswers returns detailed answers for a test result
func (db *DB) GetTestResultAnswers(resultID int) ([]*QuestionAnswerDetail, error) {
	rows, err := db.Query(`
		SELECT qa.id, qa.result_id, qa.question_id, q.question_text, qa.user_answer, q.correct_answer, qa.is_correct, q.explanation
		FROM question_answers qa
		JOIN questions q ON qa.question_id = q.id
		WHERE qa.result_id = ?
		ORDER BY qa.question_id
	`, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test result answers: %w", err)
	}
	defer rows.Close()

	var answers []*QuestionAnswerDetail
	for rows.Next() {
		answer := &QuestionAnswerDetail{}
		err := rows.Scan(&answer.ID, &answer.ResultID, &answer.QuestionID, &answer.QuestionText, &answer.UserAnswer, &answer.CorrectAnswer, &answer.IsCorrect, &answer.Explanation)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question answer: %w", err)
		}
		answers = append(answers, answer)
	}
	return answers, nil
}

// DeleteTestResult deletes a test result and its answers
func (db *DB) DeleteTestResult(resultID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete question answers first
	_, err = tx.Exec("DELETE FROM question_answers WHERE result_id = ?", resultID)
	if err != nil {
		return fmt.Errorf("failed to delete question answers: %w", err)
	}

	// Delete test result
	_, err = tx.Exec("DELETE FROM test_results WHERE id = ?", resultID)
	if err != nil {
		return fmt.Errorf("failed to delete test result: %w", err)
	}

	return tx.Commit()
}

// SaveQuestionAnswer saves a user's answer to a question
func (db *DB) SaveQuestionAnswer(resultID, questionID int, userAnswer string, isCorrect bool) error {
	_, err := db.Exec(`
		INSERT INTO question_answers (result_id, question_id, user_answer, is_correct)
		VALUES (?, ?, ?, ?)
	`, resultID, questionID, userAnswer, isCorrect)
	if err != nil {
		return fmt.Errorf("failed to save question answer: %w", err)
	}
	return nil
}

// DeleteTest deletes a test and all its associated data
func (db *DB) DeleteTest(testID int) error {
	// Start a transaction to ensure all deletions succeed or fail together
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete question answers for all questions in this test
	_, err = tx.Exec(`DELETE FROM question_answers WHERE question_id IN (SELECT id FROM questions WHERE test_id = ?)`, testID)
	if err != nil {
		return fmt.Errorf("failed to delete question answers: %w", err)
	}
	
	// Delete test results
	_, err = tx.Exec(`DELETE FROM test_results WHERE test_id = ?`, testID)
	if err != nil {
		return fmt.Errorf("failed to delete test results: %w", err)
	}
	
	// Delete questions
	_, err = tx.Exec(`DELETE FROM questions WHERE test_id = ?`, testID)
	if err != nil {
		return fmt.Errorf("failed to delete questions: %w", err)
	}
	
	// Delete the test itself
	_, err = tx.Exec(`DELETE FROM tests WHERE id = ?`, testID)
	if err != nil {
		return fmt.Errorf("failed to delete test: %w", err)
	}
	
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}