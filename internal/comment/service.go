package comment

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	CreateComment(comment entity.Comment) (*entity.Comment, *ierr.RestErr)
	FindCommentByID(id uint) (*entity.Comment, *ierr.RestErr)
	FindCommentsByCollaborator(collaboratorID uint) ([]entity.Comment, *ierr.RestErr)
	FindAllComments() ([]entity.Comment, *ierr.RestErr)
	UpdateComment(id uint, text string, authorID uint) (*entity.Comment, *ierr.RestErr)
	DeleteComment(id uint, authorID uint, authorType entity.UserType) *ierr.RestErr
}

type service struct {
	commentRepo Repository
	userRepo    user.Repository
}

func NewService(commentRepo Repository, userRepo user.Repository) Service {
	return &service{commentRepo: commentRepo, userRepo: userRepo}
}

func (s *service) CreateComment(comment entity.Comment) (*entity.Comment, *ierr.RestErr) {
	_, err := s.userRepo.FindUserByID(comment.AuthorID)
	if err != nil {
		return nil, ierr.NewBadRequestError("author not found")
	}

	_, err = s.userRepo.FindUserByID(comment.CollaboratorID)
	if err != nil {
		return nil, ierr.NewBadRequestError("collaborator not found")
	}

	if err := s.commentRepo.CreateComment(&comment); err != nil {
		return nil, ierr.NewInternalServerError("error creating comment")
	}

	return &comment, nil
}

func (s *service) FindCommentByID(id uint) (*entity.Comment, *ierr.RestErr) {
	comment, err := s.commentRepo.FindCommentByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("comment not found")
		}
		return nil, ierr.NewInternalServerError("error finding comment")
	}
	return comment, nil
}

func (s *service) FindCommentsByCollaborator(collaboratorID uint) ([]entity.Comment, *ierr.RestErr) {
	now := time.Now()
	startDate := now.AddDate(0, -3, 0)
	endDate := now.AddDate(0, 0, 1)

	comments, err := s.commentRepo.FindCommentsForUserInDateRange(collaboratorID, startDate, endDate)
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding comments for collaborator")
	}
	return comments, nil
}

func (s *service) FindAllComments() ([]entity.Comment, *ierr.RestErr) {
	comments, err := s.commentRepo.FindAllComments()
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding all comments")
	}
	return comments, nil
}

func (s *service) UpdateComment(id uint, text string, authorID uint) (*entity.Comment, *ierr.RestErr) {
	comment, restErr := s.FindCommentByID(id)
	if restErr != nil {
		return nil, restErr
	}

	if comment.AuthorID != authorID {
		return nil, ierr.NewForbiddenError("you can only edit your own comments")
	}

	comment.Text = text
	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, ierr.NewInternalServerError("error updating comment")
	}
	return comment, nil
}

func (s *service) DeleteComment(id uint, authorID uint, authorType entity.UserType) *ierr.RestErr {
	comment, restErr := s.FindCommentByID(id)
	if restErr != nil {
		return restErr
	}

	if authorType != entity.UserTypeMaster && comment.AuthorID != authorID {
		return ierr.NewForbiddenError("you do not have permission to delete this comment")
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		return ierr.NewInternalServerError("error deleting comment")
	}
	return nil
}
