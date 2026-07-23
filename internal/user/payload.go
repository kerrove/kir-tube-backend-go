package user

type UpdateProfileReq struct {
	Name     *string `json:"name"`
	Password *string `json:"password" validate:"omitempty,min=6"`
}
