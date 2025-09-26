package controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
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

	err = pcc.PollClientUsecse.SubmitVote(c, req.ID, req.Votes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
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
// @Success 200 {array} domain.PollClientResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/client/fetch [get]
func (pcc *PollClientController) Fetch(c *gin.Context) {
	id := c.Param("id")

	var polls []domain.Poll

	polls, err := pcc.PollClientUsecse.GetBySheetID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
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
	c.JSON(http.StatusOK, result)
}
