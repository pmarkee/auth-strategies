package auth

import (
	"auth-strategies/internal/common"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

const (
	success         = "success"
	jsonParseFailed = "JSON parse failed"
	emailTaken      = "Email address already taken"
)

type Api struct {
	s *Service
}

func NewApi(s *Service) *Api {
	return &Api{s}
}

// RegisterData payload for the register request
type RegisterData struct {
	Email     string `json:"email" validate:"required" example:"johndoe@example.com"`
	Password  string `json:"password" validate:"required" example:"foobar"`
	FirstName string `json:"firstName" validate:"required" example:"John"`
	LastName  string `json:"lastName" validate:"required" example:"Doe"`
}

// Register register via email and password
//
//	@Summary	register via email and password
//	@Param		request	body	RegisterData	true	"email, full name and password"
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	common.SuccessResponse
//	@Failure	400
//	@Failure	409
//	@Failure	500
//	@Router		/auth/register [post]
func (api Api) Register(w http.ResponseWriter, r *http.Request) {
	registerData := &RegisterData{}
	if err := json.NewDecoder(r.Body).Decode(registerData); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: jsonParseFailed})
		return
	}

	registerRq := &registerRq{
		email:     registerData.Email,
		password:  registerData.Password,
		firstName: registerData.FirstName,
		lastName:  registerData.LastName,
	}
	err := api.s.register(r.Context(), registerRq)
	if errors.Is(err, errEmailTaken) {
		common.WriteJSON(w, http.StatusConflict, common.ErrorResponse{Error: emailTaken})
	} else if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, common.SuccessResponse{Status: success})
}
