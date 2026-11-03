package domain

import "time"

type OrderStatus int32

const (
	OrderStatusUnspecified OrderStatus = 0
	OrderStatusPending     OrderStatus = 1
	OrderStatusPaid        OrderStatus = 2
	OrderStatusCompleted   OrderStatus = 3
	OrderStatusCanceled    OrderStatus = 4
)

type StatusHistory struct {
	Status OrderStatus `bson:"status" json:"status"`
	Note   string      `bson:"note" json:"note"`
	At     time.Time   `bson:"at" json:"at"`
}

type OrderItem struct {
	ProductID    int32   `bson:"product_id" json:"product_id"`
	Quantity     int32   `bson:"quantity" json:"quantity"`
	UnitPrice    float64 `bson:"unit_price" json:"unit_price"`
	ProductName  string  `bson:"product_name" json:"product_name"`
	ProductImage string  `bson:"product_image" json:"product_image"`
	LineTotal    float64 `bson:"line_total" json:"line_total"`
}

type Order struct {
    OrderID        int32           `bson:"order_id" json:"order_id"`
    CustomerName   string          `bson:"customer_name" json:"customer_name"`
    CustomerID     int32           `bson:"customer_id,omitempty" json:"customer_id,omitempty"` // NEW
    StaffID        string          `bson:"staff_id" json:"staff_id"`
    Items          []OrderItem     `bson:"items" json:"items"`
    VoucherCodes   []string        `bson:"voucher_codes" json:"voucher_codes"`
    TotalPrice     float64         `bson:"total_price" json:"total_price"`
    DiscountAmount float64         `bson:"discount_amount" json:"discount_amount"`
    FinalPrice     float64         `bson:"final_price" json:"final_price"`
    ShippingCost   float64         `bson:"shipping_cost" json:"shipping_cost"`
    CreatedAt      time.Time       `bson:"created_at" json:"created_at"`
    Status         OrderStatus     `bson:"status" json:"status"`
    StatusHistory  []StatusHistory `bson:"status_history" json:"status_history"`
}

