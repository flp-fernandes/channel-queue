package domain

import "time"

type ProductView struct {
	ID        int64
	ProductID int64
	ViewedAt  time.Time
}
