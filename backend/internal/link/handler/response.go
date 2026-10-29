package handler

// LinkRes is the response body for link list endpoints.
// Includes all fields the frontend needs to display and manage links.
type LinkRes struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	SortOrder   int    `json:"sortOrder"`
}

type LinkOverview struct {
	Total    int `json:"total"`
	Normal   int `json:"normal"`
	Abnormal int `json:"abnormal"`
	Pending  int `json:"pending"`
}
