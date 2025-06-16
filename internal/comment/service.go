package comment

import (
	"escala-fds-api/internal/constants"
	"escala-fds-api/internal/entity"
	"escala-fds-api/internal/user"
	"escala-fds-api/pkg/ierr"
	"strconv"

	"gorm.io/gorm"
)

type Service interface {
	CreateComment(comment entity.Comment, authorID uint, authorType entity.UserType) (*CommentResponse, *ierr.RestErr)
	FindCommentByID(id uint) (*CommentResponse, *ierr.RestErr)
	FindComments(requestorID uint, requestorType entity.UserType, filters Filters) ([]CommentResponse, *ierr.RestErr)
	UpdateComment(id uint, text string, authorID uint) (*CommentResponse, *ierr.RestErr)
	DeleteComment(id uint, requestorID uint, requestorType entity.UserType) *ierr.RestErr
}

type service struct {
	commentRepo Repository
	userRepo    user.Repository
}

func NewService(commentRepo Repository, userRepo user.Repository) Service {
	return &service{commentRepo: commentRepo, userRepo: userRepo}
}

func (s *service) CreateComment(comment entity.Comment, authorID uint, authorType entity.UserType) (*CommentResponse, *ierr.RestErr) {
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

	return s.toCommentResponse(&comment, collaborator, author)
}

func (s *service) FindCommentByID(id uint) (*CommentResponse, *ierr.RestErr) {
	comment, err := s.commentRepo.FindCommentByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("comment not found")
		}
		return nil, ierr.NewInternalServerError("error finding comment")
	}

	collaborator, err := s.userRepo.FindUserByID(comment.CollaboratorID)
	if err != nil {
		return nil, ierr.NewInternalServerError("collaborator user not found")
	}

	author, err := s.userRepo.FindUserByID(comment.AuthorID)
	if err != nil {
		return nil, ierr.NewInternalServerError("author user not found")
	}

	return s.toCommentResponse(comment, collaborator, author)
}

func (s *service) FindComments(requestorID uint, requestorType entity.UserType, filters Filters) ([]CommentResponse, *ierr.RestErr) {
	if requestorType == entity.UserTypeCollaborator {
		filters.CollaboratorID = strconv.FormatUint(uint64(requestorID), 10)
		filters.Team = ""
		filters.AuthorID = ""
	}

	comments, err := s.commentRepo.Find(filters)
	if err != nil {
		return nil, ierr.NewInternalServerError("error finding comments")
	}

	return s.buildCommentResponseList(comments)
}

func (s *service) UpdateComment(id uint, text string, authorID uint) (*CommentResponse, *ierr.RestErr) {
	comment, err := s.commentRepo.FindCommentByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ierr.NewNotFoundError("comment not found")
		}
		return nil, ierr.NewInternalServerError("error updating comment")
	}

	if comment.AuthorID != authorID {
		return nil, ierr.NewForbiddenError("you can only edit your own comments")
	}

	comment.Text = text
	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, ierr.NewInternalServerError("error updating comment")
	}

	return s.FindCommentByID(id)
}

func (s *service) DeleteComment(id uint, requestorID uint, requestorType entity.UserType) *ierr.RestErr {
	comment, err := s.commentRepo.FindCommentByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ierr.NewNotFoundError("comment not found")
		}
		return ierr.NewInternalServerError("error finding comment")
	}

	if requestorType != entity.UserTypeMaster && comment.AuthorID != requestorID {
		return ierr.NewForbiddenError("you do not have permission to delete this comment")
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		return ierr.NewInternalServerError("error deleting comment")
	}
	return nil
}

func (s *service) buildCommentResponseList(comments []entity.Comment) ([]CommentResponse, *ierr.RestErr) {
	var userIDs []uint
	userIDsSet := make(map[uint]bool)

	for _, c := range comments {
		if !userIDsSet[c.CollaboratorID] {
			userIDs = append(userIDs, c.CollaboratorID)
			userIDsSet[c.CollaboratorID] = true
		}
		if !userIDsSet[c.AuthorID] {
			userIDs = append(userIDs, c.AuthorID)
			userIDsSet[c.AuthorID] = true
		}
	}

	users, err := s.userRepo.FindUsersByIDs(userIDs)
	if err != nil {
		return nil, ierr.NewInternalServerError("error fetching user data for comments")
	}

	userMap := make(map[uint]*entity.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	var responses []CommentResponse
	for _, c := range comments {
		collaborator := userMap[c.CollaboratorID]
		author := userMap[c.AuthorID]
		if collaborator != nil && author != nil {
			resp, _ := s.toCommentResponse(&c, collaborator, author)
			responses = append(responses, *resp)
		}
	}
	return responses, nil
}

func (s *service) toCommentResponse(comment *entity.Comment, collaborator, author *entity.User) (*CommentResponse, *ierr.RestErr) {
	response := &CommentResponse{
		ID:           comment.ID,
		Collaborator: user.ToUserResponse(collaborator),
		Author:       user.ToUserResponse(author),
		Text:         comment.Text,
		Date:         comment.Date.Format(constants.ApiDateLayout),
		CreatedAt:    comment.CreatedAt.Format(constants.ApiTimestampLayout),
		UpdatedAt:    comment.UpdatedAt.Format(constants.ApiTimestampLayout),
	}
	return response, nil
}
