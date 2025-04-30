package auth

import (
	"auth-strategies/internal/common"
	"encoding/json"
	"errors"
	"github.com/alexedwards/scs/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const (
	success         = "Success"
	jsonParseFailed = "JSON parse failed"
	emailTaken      = "Email address already taken"
)

type Api struct {
	s            *Service
	sessionStore *scs.SessionManager
	hmacSecret   []byte
}

func NewApi(s *Service, sessionStore *scs.SessionManager, hmacSecret []byte) *Api {
	return &Api{s, sessionStore, hmacSecret}
}

// RegisterData payload for the register request
type RegisterData struct {
	Email     string `json:"email" validate:"required" example:"johndoe@example.com"`
	Password  string `json:"password" validate:"required" example:"foobar"`
	FirstName string `json:"firstName" validate:"required" example:"John"`
	LastName  string `json:"lastName" validate:"required" example:"Doe"`
}

// LoginData payload for login
type LoginData struct {
	Email    string `json:"email" validate:"required" example:"johndoe@example.com"`
	Password string `json:"password" validate:"required" example:"foobar"`
}

// AccessTokenResponse response containing the generated JWT access token
type AccessTokenResponse struct {
	AccessToken string `json:"accessToken" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJIUzI1NiIsImV4cCI6MTc0NjAyMTY0NSwic3ViIjoiMDllMjNjNDAtM2JjMC00OTI0LWIxMDAtMmI3YjMyZDMxMGZlIn0.2ZSXXjxsLqeQOaovDU4tuj-8-6Hd7pUBxLpchURpWDU"`
}

// ApiKeyResponse response containing the generated API key
type ApiKeyResponse struct {
	// ApiKey is a string formatted as "xxx.yyy" where "xxx" is a public id, and "yyy" is a secret that is only stored encrypted on the server
	ApiKey string `json:"apiKey" validate:"required" example:"fa40d13983db9cf8a19477d42f652726.37c476287cb99a1e6b1ad69006ad8c48d7c494368a21e16e5dbd2d29235de87b"`
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
func (api *Api) Register(w http.ResponseWriter, r *http.Request) {
	registerData := &RegisterData{}
	if err := json.NewDecoder(r.Body).Decode(registerData); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: jsonParseFailed})
		return
	}

	rq := &registerRq{
		email:     registerData.Email,
		password:  registerData.Password,
		firstName: registerData.FirstName,
		lastName:  registerData.LastName,
	}
	err := api.s.register(r.Context(), rq)
	if errors.Is(err, errEmailTaken) {
		common.WriteJSON(w, http.StatusConflict, common.ErrorResponse{Error: emailTaken})
	} else if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, common.SuccessResponse{Status: success})
}

// Login login via email and password
//
//	@Summary	login via email and password
//	@Param		request	body	LoginData	true	"email and password"
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	common.SuccessResponse
//	@Failure	401
//	@Failure	500
//	@Header		200			{string}	Set-Cookie	"Session cookie"
//	@Router		/auth/login	[post]
func (api *Api) Login(w http.ResponseWriter, r *http.Request) {
	id := api.loginHelper(w, r)
	if id == nil {
		return
	}

	api.sessionStore.Put(r.Context(), "user_id", id.String())
	common.WriteJSON(w, http.StatusOK, common.SuccessResponse{Status: success})
}

// LoginToken exchange email and password for an access and refresh token
//
//	@Summary	exchange email and password for an access and refresh token
//	@Param		request	body	LoginData	true	"email and password"
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	AccessTokenResponse
//	@Failure	401
//	@Failure	500
//	@Router		/auth/token/login	[post]
func (api *Api) LoginToken(w http.ResponseWriter, r *http.Request) {
	id := api.loginHelper(w, r)
	if id == nil {
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id.String(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"alg": jwt.SigningMethodHS256.Alg(),
	})
	tokenString, err := token.SignedString(api.hmacSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to sign access token")
		return
	}

	common.WriteJSON(w, http.StatusOK, AccessTokenResponse{AccessToken: tokenString})
}

func (api *Api) loginHelper(w http.ResponseWriter, r *http.Request) *uuid.UUID {
	loginData := &LoginData{}
	if err := json.NewDecoder(r.Body).Decode(loginData); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: jsonParseFailed})
		return nil
	}

	id, err := api.s.checkPassword(r.Context(), loginData.Email, loginData.Password)
	if errors.Is(err, errInvalidCredentials) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("login failed")
		return nil
	}

	return id
}

// Logout log the user out of the current session
//
//	@Summary	log the user out of the current session
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	common.SuccessResponse
//	@Failure	500
//	@Router		/auth/logout	[post]
func (api *Api) Logout(w http.ResponseWriter, r *http.Request) {
	err := api.sessionStore.Destroy(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to destroy session")
		return
	}
	common.WriteJSON(w, http.StatusOK, common.SuccessResponse{Status: success})
}

// GenerateApiKey generate an API key for the authenticated user
//
//	@Summary		generate an API key for the authenticated user
//	@Description	generate an API key for the authenticated user (WARNING: the key will only be returned once and cannot be retrieved later!)
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	ApiKeyResponse
//	@Failure		401
//	@Failure		500
//	@Router			/auth/api-key [get]
//	@Security		session
func (api *Api) GenerateApiKey(w http.ResponseWriter, r *http.Request) {
	id := common.GetUserIdFromContext(w, r)
	if id == nil {
		return
	}

	key, err := api.s.generateApiKey(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("failed to generate api key")
		return
	}

	common.WriteJSON(w, http.StatusOK, ApiKeyResponse{ApiKey: key})
}
