package holiday

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
	holidayRoutes := router.Group("/holidays")
	holidayRoutes.Use(auth.Middleware())
	{
		holidayRoutes.POST("", h.Create)
		holidayRoutes.GET("", h.FindAll)
		holidayRoutes.GET("/:id", h.FindByID)
		holidayRoutes.PUT("/:id", h.Update)
		holidayRoutes.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid date format, use YYYY-MM-DD"))
		return
	}

	holiday := entity.Holiday{Name: req.Name, Date: date, Type: req.Type}
	newHoliday, errSvc := h.service.CreateHoliday(holiday)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	c.JSON(http.StatusCreated, ToHolidayResponse(newHoliday))
}

func (h *Handler) FindAll(c *gin.Context) {
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	if startDateStr == "" || endDateStr == "" {
		holidays, err := h.service.FindAllHolidays()
		if err != nil {
			c.JSON(err.Code, err)
			return
		}
		var res []HolidayResponse
		for _, holiday := range holidays {
			res = append(res, ToHolidayResponse(&holiday))
		}
		c.JSON(http.StatusOK, res)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid startDate format"))
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid endDate format"))
		return
	}

	holidays, errSvc := h.service.FindHolidaysByDateRange(startDate, endDate)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	var res []HolidayResponse
	for _, holiday := range holidays {
		res = append(res, ToHolidayResponse(&holiday))
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) FindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	holiday, err := h.service.FindHolidayByID(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.JSON(http.StatusOK, ToHolidayResponse(holiday))
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req UpdateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError(err.Error()))
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, ierr.NewBadRequestError("invalid date format, use YYYY-MM-DD"))
		return
	}

	holidayData := entity.Holiday{Name: req.Name, Date: date, Type: req.Type}
	updatedHoliday, errSvc := h.service.UpdateHoliday(uint(id), holidayData)
	if errSvc != nil {
		c.JSON(errSvc.Code, errSvc)
		return
	}
	c.JSON(http.StatusOK, ToHolidayResponse(updatedHoliday))
}

func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.service.DeleteHoliday(uint(id))
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	c.Status(http.StatusNoContent)
}
