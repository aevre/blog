package handler

type GetCaptchaReq struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"omitempty,oneof=Register PasswordReset ChangeEmail"`
}

type RegisterReq struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=8,max=64,eqfield=Password"`
	Captcha         string `json:"captcha" validate:"required"`
}

type LoginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}
