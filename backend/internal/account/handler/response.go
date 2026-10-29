package handler

// LoginRes is the response body for login/register endpoints.
// Role remains a string to keep the HTTP contract independent of domain types.
type LoginRes struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"-"`
	UUID         string  `json:"uuid"`
	Avatar       *string `json:"avatar"`
	Username     string  `json:"username"`
	Role         string  `json:"role"`
}

type RefreshRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"-"`
}
