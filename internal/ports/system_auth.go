package ports

type SystemAuthService interface {
	Login(username, password string) (*TokenPair, error)
	Refresh(refreshToken string) (*TokenPair, error)
}
