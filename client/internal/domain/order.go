package domain

type Order struct {
	ID        string
	UserID    string
	MarketID  string
	OrderType string
	Status    string
	Price     float64
	Quantity  float64
}

func NewOrder(userId string) *Order {
	return &Order{
		UserID: userId,
	}
}
