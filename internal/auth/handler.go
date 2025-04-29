package auth

type Api struct {
	s *Service
}

func NewApi(s *Service) *Api {
	return &Api{s}
}
