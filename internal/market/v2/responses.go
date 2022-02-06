package market

type getItemsResponse struct {
	Success  bool                  `json:"success,omitempty"`
	Currency string                `json:"currency,omitempty"`
	Data     map[string][]itemData `json:"data,omitempty"`
}

type itemData struct {
	Price string `json:"price,omitempty"`
}
