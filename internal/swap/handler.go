package swap

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
	swapRoutes := router.Group("/swaps")
	swapRoutes.Use(auth.Middleware())
	{
		swapRoutes.POST("", h.Create)
		swapRoutes.GET("", h.FindAll)
		swapRoutes.GET("/user/:id", h.FindByUser)
		swapRoutes.GET("/:id", h.FindByID)
		swapRoutes.PATCH("/:id/status", h.UpdateStatus)
		swapRoutes.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	requesterID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	originalDate, err := time.ParseInLocation("2006-01-02", req.OriginalDate, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid originalDate format"))
		return
	}
	newDate, err := time.ParseInLocation("2006-01-02", req.NewDate, time.UTC)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid newDate format"))
		return
	}

	swapEntity := entity.Swap{
		InvolvedCollaboratorID: req.InvolvedCollaboratorID,
		OriginalDate:           originalDate,
		NewDate:                newDate,
		OriginalShift:          req.OriginalShift,
		NewShift:               req.NewShift,
		Reason:                 req.Reason,
	}

	newSwap, errSvc := h.service.CreateSwap(swapEntity, requesterID)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	c.JSON(http.StatusCreated, ToSwapResponse(newSwap))
}

func (h *Handler) FindAll(c *gin.Context) {
	swaps, err := h.service.FindAllSwaps()
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	var res []SwapResponse
	for _, swap := range swaps {
		res = append(res, ToSwapResponse(&swap))
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) FindByUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	statusFilter := c.Query("status")

	swaps, err := h.service.FindSwapsForUser(uint(id), statusFilter)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	var res []SwapResponse
	for _, swap := range swaps {
		res = append(res, ToSwapResponse(&swap))
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	swap, err := h.service.FindSwapByID(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToSwapResponse(swap))
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	approverID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}
	var req UpdateSwapStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}
	updatedSwap, err := h.service.ApproveOrRejectSwap(uint(id), approverID, req.Status)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToSwapResponse(updatedSwap))
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requesterID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}
	requesterType, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	err := h.service.DeleteSwap(uint(id), requesterID, entity.UserType(requesterType))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.Status(http.StatusNoContent)
}
