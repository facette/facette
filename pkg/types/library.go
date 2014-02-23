package types

// ItemResponse represents an item response struct in the server library.
type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Modified    string `json:"modified"`
}

// ItemListResponse represents an items list response struct in the server library.
type ItemListResponse struct {
	Items []*ItemResponse `json:"items"`
}

func (response ItemListResponse) Len() int {
	return len(response.Items)
}

func (response ItemListResponse) Less(i, j int) bool {
	return response.Items[i].Name < response.Items[j].Name
}

func (response ItemListResponse) Swap(i, j int) {
	response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
}
