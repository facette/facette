package server

// ItemResponse represents an item response structure in the server backend.
type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Modified    string `json:"modified"`
}

// ItemListResponse represents a list of items response structure in the backend server.
type ItemListResponse []*ItemResponse

func (r ItemListResponse) Len() int {
	return len(r)
}

func (r ItemListResponse) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r ItemListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
