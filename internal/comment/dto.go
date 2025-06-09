package comment

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"time"
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
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

func ToCommentResponse(comment *entity.Comment) CommentResponse {
	return CommentResponse{
		ID:           comment.ID,
		Collaborator: user.ToUserResponse(&comment.Collaborator),
		Author:       user.ToUserResponse(&comment.Author),
		Text:         comment.Text,
		Date:         comment.Date.Format("2006-01-02"),
		CreatedAt:    comment.CreatedAt,
		UpdatedAt:    comment.UpdatedAt,
	}
}
