package db

import (
	"bugtracker-backend/internal/models"
	"fmt"
	"strconv"
	"time"
)

func CreateComment(bugID string, comment *models.Comment) error {
	bugIDInt, err := strconv.Atoi(bugID)
	if err != nil {
		return fmt.Errorf("invalid bug ID format")
	}

	_, err = GetBug(bugIDInt)
	if err != nil {
		return fmt.Errorf("bug not found")
	}

	comment.BugID = bugIDInt
	comment.CreatedAt = time.Now()

	query := `
		INSERT INTO comments (bug_id, content, author, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err = DB.QueryRow(
		query,
		comment.BugID,
		comment.Content,
		comment.Author,
		comment.CreatedAt,
	).Scan(&comment.ID)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func GetComments(bugID string) ([]models.Comment, error) {
	bugIDInt, err := strconv.Atoi(bugID)
	if err != nil {
		return nil, fmt.Errorf("invalid bug ID format")
	}

	_, err = GetBug(bugIDInt)
	if err != nil {
		return nil, fmt.Errorf("bug not found")
	}

	query := `
		SELECT id, bug_id, content, author, created_at
		FROM comments
		WHERE bug_id = $1
		ORDER BY created_at ASC;
	`

	rows, err := DB.Query(query, bugIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment

	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.BugID,
			&comment.Content,
			&comment.Author,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}