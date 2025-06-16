package user

import (
	"escala-fds-api/internal/auth"
	"escala-fds-api/internal/entity"
	"escala-fds-api/pkg/ierr"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)

	userRoutes := router.Group("/users")
	userRoutes.Use(auth.Middleware())
	{
		userRoutes.POST("", h.CreateUser)
		userRoutes.GET("", h.FindAll)
		userRoutes.GET("/:id", h.FindByID)
		userRoutes.PUT("/:id/personal", h.UpdatePersonalData)
		userRoutes.PUT("/:id/work", h.UpdateWorkData)
		userRoutes.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := ierr.NewBadRequestValidationError("invalid request body", nil)
		c.JSON(restErr.Code, restErr)
		return
	}

	token, user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  ToUserResponse(user),
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	creatorTypeStr, errAuth := auth.GetUserTypeFromContext(c)
	if errAuth != nil {
		c.JSON(errAuth.Code, errAuth)
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := ierr.NewBadRequestValidationError("invalid request body", nil)
		c.JSON(restErr.Code, restErr)
		return
	}

	userEntity := entity.User{
		Email:             req.Email,
		Password:          req.Password,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		PhoneNumber:       req.PhoneNumber,
		UserType:          req.UserType,
		Team:              req.Team,
		Position:          req.Position,
		Shift:             req.Shift,
		WeekdayOff:        req.WeekdayOff,
		InitialWeekendOff: req.InitialWeekendOff,
		SuperiorID:        req.SuperiorID,
	}

	newUser, err := h.service.CreateUser(userEntity, entity.UserType(creatorTypeStr))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(http.StatusCreated, ToUserResponse(newUser))
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := h.service.FindUserByID(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToUserResponse(user))
}

func (h *Handler) FindAll(c *gin.Context) {
	requestorTypeStr, _ := auth.GetUserTypeFromContext(c)
	requestorTeamStr, _ := auth.GetUserTeamFromContext(c)

	users, err := h.service.FindAllUsers(entity.UserType(requestorTypeStr), entity.TeamName(requestorTeamStr))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, ToUserResponse(&user))
	}
	c.JSON(http.StatusOK, userResponses)
}

func (h *Handler) UpdatePersonalData(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requestorId, _ := auth.GetUserIDFromContext(c)
	requestorType, _ := auth.GetUserTypeFromContext(c)

	var req UpdatePersonalDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := ierr.NewBadRequestValidationError("invalid request body", nil)
		c.JSON(restErr.Code, restErr)
		return
	}

	userEntity := entity.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
	}

	updatedUser, err := h.service.UpdatePersonalData(uint(id), requestorId, entity.UserType(requestorType), userEntity)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToUserResponse(updatedUser))
}

func (h *Handler) UpdateWorkData(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requestorType, _ := auth.GetUserTypeFromContext(c)

	var req UpdateWorkDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		restErr := ierr.NewBadRequestValidationError("invalid request body", nil)
		c.JSON(restErr.Code, restErr)
		return
	}

	userEntity := entity.User{
		Team:              req.Team,
		Position:          req.Position,
		Shift:             req.Shift,
		WeekdayOff:        req.WeekdayOff,
		InitialWeekendOff: req.InitialWeekendOff,
		SuperiorID:        req.SuperiorID,
	}

	updatedUser, err := h.service.UpdateWorkData(uint(id), entity.UserType(requestorType), userEntity)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToUserResponse(updatedUser))
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	requestorType, _ := auth.GetUserTypeFromContext(c)

	err := h.service.DeleteUser(uint(id), entity.UserType(requestorType))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.Status(http.StatusNoContent)
}
