package types

// CollectionResponse represents a collection response struct in the server library backend.
type CollectionResponse struct {
	ItemResponse
	Parent      *string `json:"parent"`
	HasChildren bool    `json:"has_children"`
}

// CollectionListResponse represents a collections list response struct in the server library backend.
type CollectionListResponse struct {
	Items []*CollectionResponse `json:"items"`
}

func (response CollectionListResponse) Len() int {
	return len(response.Items)
}

func (response CollectionListResponse) Less(i, j int) bool {
	return response.Items[i].Name < response.Items[j].Name
}

func (response CollectionListResponse) Swap(i, j int) {
	response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
}
