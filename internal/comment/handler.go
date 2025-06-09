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
		commentRoutes.GET("", h.FindAll)
		commentRoutes.GET("/user/:id", h.FindByCollaborator)
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

	date, err := time.ParseInLocation("2006-01-02", req.Date, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid date format, use YYYY-MM-DD"))
		return
	}

	commentEntity := entity.Comment{
		CollaboratorID: req.CollaboratorID,
		AuthorID:       authorID,
		Text:           req.Text,
		Date:           date,
	}

	newComment, errSvc := h.service.CreateComment(commentEntity)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	c.JSON(http.StatusCreated, ToCommentResponse(newComment))
}

func (h *Handler) FindAll(c *gin.Context) {
	comments, err := h.service.FindAllComments()
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	var res []CommentResponse
	for _, comment := range comments {
		res = append(res, ToCommentResponse(&comment))
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) FindByCollaborator(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	comments, err := h.service.FindCommentsByCollaborator(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	var res []CommentResponse
	for _, comment := range comments {
		res = append(res, ToCommentResponse(&comment))
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	comment, err := h.service.FindCommentByID(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToCommentResponse(comment))
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
	c.JSON(http.StatusOK, ToCommentResponse(updatedComment))
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
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

	err := h.service.DeleteComment(uint(id), authorID, entity.UserType(authorTypeStr))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.Status(http.StatusNoContent)
}
