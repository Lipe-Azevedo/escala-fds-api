package certificate

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
	routes := router.Group("/certificates")
	routes.Use(auth.Middleware())
	{
		routes.POST("", h.Create)
		routes.GET("", h.FindAll)
		routes.GET("/user/:id", h.FindByUser)
		routes.PATCH("/:id/status", h.UpdateStatus)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	collaboratorID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid startDate format"))
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid endDate format"))
		return
	}

	cert := entity.Certificate{
		CollaboratorID: collaboratorID,
		StartDate:      startDate,
		EndDate:        endDate,
		Reason:         req.Reason,
	}

	newCert, errSvc := h.service.CreateCertificate(cert)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}

	c.JSON(http.StatusCreated, ToResponse(newCert))
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	userType, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}
	if userType != string(entity.UserTypeMaster) {
		c.JSON(http.StatusForbidden, ierr.NewForbiddenError("only masters can approve or reject certificates"))
		return
	}

	approverID, errAuth := auth.GetUserIDFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	updatedCert, errSvc := h.service.ApproveOrReject(uint(id), approverID, req.Status)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}

	c.JSON(http.StatusOK, ToResponse(updatedCert))
}

func (h *Handler) FindAll(c *gin.Context) {
	userType, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}
	if userType != string(entity.UserTypeMaster) {
		c.JSON(http.StatusForbidden, ierr.NewForbiddenError("only masters can view all certificates"))
		return
	}

	certs, errSvc := h.service.FindAll()
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}

	var response []CertificateResponse
	for _, cert := range certs {
		response = append(response, ToResponse(&cert))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) FindByUser(c *gin.Context) {
	collaboratorID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	requestorID, _ := auth.GetUserIDFromContext(c)
	requestorType, _ := auth.GetUserTypeFromContext(c)

	if requestorType != string(entity.UserTypeMaster) && uint(collaboratorID) != requestorID {
		c.JSON(http.StatusForbidden, ierr.NewForbiddenError("you can only view your own certificates"))
		return
	}

	certs, errSvc := h.service.FindByCollaborator(uint(collaboratorID))
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}

	var response []CertificateResponse
	for _, cert := range certs {
		response = append(response, ToResponse(&cert))
	}
	c.JSON(http.StatusOK, response)
}
