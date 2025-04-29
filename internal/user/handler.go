package user

import (
	"auth-strategies/internal/utils"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"net/http"
)

type Api struct {
	s *Service
}

func NewApi(s *Service) *Api {
	return &Api{s}
}

// GetUserInfoResponse contains the first name and last name of a user
type GetUserInfoResponse struct {
	FirstName string `json:"firstName" validate:"required" example:"John"`
	LastName  string `json:"lastName" validate:"required" example:"Doe"`
}

// GetUserInfo fetch a user's first and last name based on their id
//
//	@Summary	fetch a user's first and last name based on their id
//	@Tags		user
//	@Param		id	query	string	true	"id of the user"
//	@Produce	json
//	@Success	200	{object}	GetUserInfoResponse
//	@Failure	400	{object}	utils.ErrorResponse
//	@Failure	401	{object}	utils.ErrorResponse
//	@Failure	500
//	@Router		/user [get]
func (api *Api) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "missing id"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "invalid id"})
		return
	}

	userData, err := api.s.getUserData(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "user not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, GetUserInfoResponse{userData.FirstName, userData.LastName})
}
