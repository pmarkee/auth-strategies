package user

import (
	"auth-strategies/internal/common"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

// GetUserInfoBasic fetch the authenticated user's first and last name - basic auth
//
//	@Summary	fetch the authenticated user's first and last name - basic auth
//	@Tags		user
//	@Produce	json
//	@Success	200	{object}	GetUserInfoResponse
//	@Failure	400	{object}	common.ErrorResponse
//	@Failure	401	{object}	common.ErrorResponse
//	@Failure	500
//	@Router		/user/basic [get]
//	@Security	BasicAuth
func (api *Api) GetUserInfoBasic(w http.ResponseWriter, r *http.Request) {
	api.getUserInfo(w, r)
}

// GetUserInfoSession fetch the authenticated user's first and last name - session auth
//
//	@Summary	fetch the authenticated user's first and last name - session auth
//	@Tags		user
//	@Produce	json
//	@Success	200	{object}	GetUserInfoResponse
//	@Failure	400	{object}	common.ErrorResponse
//	@Failure	401	{object}	common.ErrorResponse
//	@Failure	500
//	@Router		/user/session [get]
//	@Security	session
func (api *Api) GetUserInfoSession(w http.ResponseWriter, r *http.Request) {
	api.getUserInfo(w, r)
}

// GetUserInfoToken fetch the authenticated user's first and last name - token auth
//
//	@Summary	fetch the authenticated user's first and last name - token auth
//	@Tags		user
//	@Produce	json
//	@Success	200	{object}	GetUserInfoResponse
//	@Failure	400	{object}	common.ErrorResponse
//	@Failure	401	{object}	common.ErrorResponse
//	@Failure	500
//	@Router		/user/token [get]
//	@Security	Bearer
func (api *Api) GetUserInfoToken(w http.ResponseWriter, r *http.Request) {
	api.getUserInfo(w, r)
}

func (api *Api) getUserInfo(w http.ResponseWriter, r *http.Request) {
	id := getUserIdFromContext(w, r)
	if id == nil {
		return
	}

	userData, err := api.s.getUserData(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		common.WriteJSON(w, http.StatusNotFound, common.ErrorResponse{Error: "user not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	common.WriteJSON(w, http.StatusOK, GetUserInfoResponse{userData.FirstName, userData.LastName})
}

func getUserIdFromContext(w http.ResponseWriter, r *http.Request) *uuid.UUID {
	id, ok := r.Context().Value("id").(*uuid.UUID)
	if !ok {
		// Should not be reached
		log.Error().Msg("failed to read user id from context")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return id
}
