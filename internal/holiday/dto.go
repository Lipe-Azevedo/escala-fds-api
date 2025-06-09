package holiday

import (
	"escala-fds-api/internal/constants"
	"escala-fds-api/internal/entity"
)

type CreateHolidayRequest struct {
	Name string             `json:"name" binding:"required"`
	Date string             `json:"date" binding:"required"`
	Type entity.HolidayType `json:"type" binding:"required"`
}

type UpdateHolidayRequest struct {
	Name string             `json:"name" binding:"required"`
	Date string             `json:"date" binding:"required"`
	Type entity.HolidayType `json:"type" binding:"required"`
}

type HolidayResponse struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Date      string             `json:"date"`
	Type      entity.HolidayType `json:"type"`
	CreatedAt string             `json:"createdAt"`
}

func ToHolidayResponse(holiday *entity.Holiday) HolidayResponse {
	return HolidayResponse{
		ID:        holiday.ID,
		Name:      holiday.Name,
		Date:      holiday.Date.Format("2006-01-02"),
		Type:      holiday.Type,
		CreatedAt: holiday.CreatedAt.Format(constants.ApiTimestampLayout),
	}
}
