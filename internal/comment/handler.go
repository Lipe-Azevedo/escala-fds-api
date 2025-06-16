package comment

import (
	"escala-fds-api/internal/auth"
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	commentRoutes := router.Group("/comments")
	commentRoutes.Use(auth.Middleware())
	{
		commentRoutes.POST("", h.Create)
		commentRoutes.GET("", h.Find)
		commentRoutes.GET("/:id", h.FindByID)
		commentRoutes.PUT("/:id", h.Update)
		commentRoutes.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	authorID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	authorTypeStr, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	date, err := time.ParseInLocation("2006-01-02", req.Date, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid date format, use yyyy-MM-dd"))
		return
	}

	commentEntity := entity.Comment{
		CollaboratorID: req.CollaboratorID,
		Text:           req.Text,
		Date:           date,
	}

	newComment, errSvc := h.service.CreateComment(commentEntity, authorID, entity.UserType(authorTypeStr))
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	c.JSON(http.StatusCreated, newComment)
}

func (h *Handler) Find(c *gin.Context) {
	requestorID, _ := auth.GetUserIDFromContext(c)
	requestorType, _ := auth.GetUserTypeFromContext(c)

	filters := Filters{
		StartDate:      c.Query("startDate"),
		EndDate:        c.Query("endDate"),
		CollaboratorID: c.Query("collaboratorId"),
		AuthorID:       c.Query("authorId"),
		Team:           c.Query("team"),
	}

	comments, err := h.service.FindComments(requestorID, entity.UserType(requestorType), filters)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	comment, err := h.service.FindCommentByID(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, comment)
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	authorID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	updatedComment, err := h.service.UpdateComment(uint(id), req.Text, authorID)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, updatedComment)
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requestorID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}
	requestorTypeStr, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	err := h.service.DeleteComment(uint(id), requestorID, entity.UserType(requestorTypeStr))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.Status(http.StatusNoContent)
}
