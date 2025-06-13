package user

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}
