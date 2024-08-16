package types

type GetObjectRequest struct {
	Name string `json:"name"`
}

type GetObjectResponse struct {
	Object Object `json:"object"`
}
