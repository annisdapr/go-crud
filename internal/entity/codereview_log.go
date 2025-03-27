package entity

import "time"

type CodeReviewLog struct {
	ID           int       `json:"id"`
	RepositoryID int       `json:"repository_id"`
	ReviewResult string    `json:"review_result"`
	CreatedAt    time.Time `json:"created_at"`
}
