package comment

import (
	"escala-fds-api/internal/user"
)

type CreateCommentRequest struct {
	CollaboratorID uint   `json:"collaboratorId" binding:"required"`
	Text           string `json:"text" binding:"required"`
	Date           string `json:"date" binding:"required"`
}

type UpdateCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

type CommentResponse struct {
	ID           uint              `json:"id"`
	Collaborator user.UserResponse `json:"collaborator"`
	Author       user.UserResponse `json:"author"`
	Text         string            `json:"text"`
	Date         string            `json:"date"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type Filters struct {
	StartDate      string
	EndDate        string
	CollaboratorID string
	AuthorID       string
	Team           string
}
