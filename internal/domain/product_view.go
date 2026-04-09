package domain

import "time"

type ProductView struct {
	ID        int64
	ProductID string
	ViewedAt  time.Time
}
