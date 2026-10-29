package handler

type InitializeReq struct {
	Username        string `json:"username" validate:"required,min=3,max=32"`
	Email           string `json:"email" validate:"required,email,max=255"`
	Password        string `json:"password" validate:"required,min=8,max=128"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
	SiteName        string `json:"siteName" validate:"required,min=1,max=100"`
	Logo            string `json:"logo" validate:"omitempty,url,max=2048"`
	Greeting        string `json:"greeting" validate:"omitempty,max=255"`
	Description     string `json:"description" validate:"omitempty,max=1000"`
}
