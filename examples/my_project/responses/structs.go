package responses

import "examples/my_project/models"

// simplegen:sort-by-keys -type *examples/my_project/models.User
type UsersResponse struct {
	Users     []*models.User
	RequestID string
}

// simplegen:sort-by-keys -type *examples/my_project/models.User -suffix ByEmail -fieldName Email -fieldType string
type UsersResponseByEmail struct {
	Users     []*models.User
	RequestID string
}
