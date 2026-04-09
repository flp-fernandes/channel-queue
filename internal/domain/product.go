package domain

import "time"

type Product struct {
	ID        int64
	SKU       string
	Name      string
	PriceCent int64
	CreatedAt time.Time
}
