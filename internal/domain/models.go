package domain

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	Image    Image   `json:"image"`
}

type Image struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}

type OrderItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type OrderRequest struct {
	CouponCode string             `json:"couponCode,omitempty"`
	Items      []OrderItemRequest `json:"items"`
}

type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	ID        string      `json:"id"`
	Total     float64     `json:"total"`
	Discounts float64     `json:"discounts"`
	Items     []OrderItem `json:"items"`
	Products  []Product   `json:"products"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}
