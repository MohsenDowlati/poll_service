package controller

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PollClientController struct {
	PollClientUsecse domain.PollClientUsecase
}

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
