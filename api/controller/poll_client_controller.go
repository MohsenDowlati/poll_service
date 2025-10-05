package controller

import (
	"errors"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type PollClientController struct {
	PollClientUsecse domain.PollClientUsecase
}

// Submit records votes for a poll.
// @Summary Submit poll votes
// @Description Submit votes for a poll.
// @Tags Polls
// @Accept json
// @Produce json
// @Param payload body domain.PollClientRequest true "Votes payload"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/submit [post]
func (pcc *PollClientController) Submit(c *gin.Context) {
	var req domain.PollClientRequest

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	err = pcc.PollClientUsecse.SubmitVote(c, req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNoVotesSubmitted) || errors.Is(err, domain.ErrNoOpinionSubmitted) {
			status = http.StatusBadRequest
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "Votes are submitted."})
}

// Fetch returns polls for a sheet.
// @Summary Get polls for sheet
// @Description Retrieve published polls for a sheet.
// @Tags Polls
// @Produce json
// @Param id query string true "Sheet identifier"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.PollClientListResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/client/fetch [get]
func (pcc *PollClientController) Fetch(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		id = c.Param("id")
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "sheet id is required"})
		return
	}

	pagination := extractPagination(c)

	polls, total, err := pcc.PollClientUsecse.GetBySheetID(c, id, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	sheet, err := pcc.PollClientUsecse.GetSheet(c, id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		c.JSON(status, domain.ErrorResponse{Message: err.Error()})
		return
	}

	var result []domain.PollClientResponse

	for _, poll := range polls {
		result = append(result, domain.PollClientResponse{
			ID:          poll.ID.Hex(),
			Title:       poll.Title,
			Options:     poll.Options,
			PollType:    poll.PollType,
			Description: poll.Description,
		})
	}

	response := domain.PollClientListResponse{
		Data: result,
		Sheet: domain.PollClientSheetMeta{
			ID:              sheet.ID.Hex(),
			Title:           sheet.Title,
			IsPhoneRequired: sheet.IsPhoneRequired,
		},
		Pagination: domain.NewPaginationResult(pagination, total),
	}

	c.JSON(http.StatusOK, response)
}
