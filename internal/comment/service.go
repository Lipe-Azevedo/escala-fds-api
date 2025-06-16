package comment

import (
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"strconv"

	"gorm.io/gorm"
)

type Service interface {
	CreateComment(comment entity.Comment, authorID uint, authorType entity.UserType) (*entity.Comment, *ierr.RestErr)
	FindCommentByID(id uint) (*entity.Comment, *ierr.RestErr)
	FindComments(requestorID uint, requestorType entity.UserType, filters Filters) ([]entity.Comment, *ierr.RestErr)
	UpdateComment(id uint, text string, authorID uint) (*entity.Comment, *ierr.RestErr)
	DeleteComment(id uint, requestorID uint, requestorType entity.UserType) *ierr.RestErr
}

type service struct {
	commentRepo Repository
	userRepo    user.Repository
}

func NewService(commentRepo Repository, userRepo user.Repository) Service {
	return &service{commentRepo: commentRepo, userRepo: userRepo}
}

func (s *service) CreateComment(comment entity.Comment, authorID uint, authorType entity.UserType) (*entity.Comment, *ierr.RestErr) {
	author, err := s.userRepo.FindUserByID(authorID)
	if err != nil {
		return nil, ierr.NewBadRequestError("author not found")
	}

	collaborator, err := s.userRepo.FindUserByID(comment.CollaboratorID)
	if err != nil {
		return nil, ierr.NewBadRequestError("collaborator not found")
	}

	isSuperior := (collaborator.SuperiorID != nil && *collaborator.SuperiorID == authorID)
	if author.UserType != entity.UserTypeMaster && !isSuperior {
		return nil, ierr.NewForbiddenError("only masters or direct superiors can add comments")
	}

	comment.AuthorID = authorID
	if err := s.commentRepo.CreateComment(&comment); err != nil {
		return nil, ierr.NewInternalServerError("error creating comment")
	}

	newComment, err := s.commentRepo.FindCommentByID(comment.ID)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching newly created comment")
	}

	return newComment, nil
}

func (s *service) FindComments(requestorID uint, requestorType entity.UserType, filters Filters) ([]entity.Comment, *ierr.RestErr) {
	if requestorType == entity.UserTypeCollaborator {
		filters.CollaboratorID = strconv.FormatUint(uint64(requestorID), 10)
		filters.Team = ""
		filters.AuthorID = ""
	}

	comments, err := s.commentRepo.Find(filters)
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding comments")
	}
	return comments, nil
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

func (s *service) DeleteComment(id uint, requestorID uint, requestorType entity.UserType) *ierr.RestErr {
	comment, restErr := s.FindCommentByID(id)
	if restErr != nil {
		return restErr
	}

	if requestorType != entity.UserTypeMaster && comment.AuthorID != requestorID {
		return ierr.NewForbiddenError("you do not have permission to delete this comment")
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		return ierr.NewInternalServerError("error deleting comment")
	}
	return nil
}
