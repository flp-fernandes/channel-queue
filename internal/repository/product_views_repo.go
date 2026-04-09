package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/flp-fernandes/product-views/internal/domain"
)

type ProductViewsRepository struct {
	db *sql.DB
}

func NewProductViewsRepository(db *sql.DB) *ProductViewsRepository {
	return &ProductViewsRepository{db: db}
}

func (r *ProductViewsRepository) BulkInsert(ctx context.Context, views []domain.ProductView) error {
	if len(views) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO product_views (product_id, viewed_at) VALUES ")

	args := make([]interface{}, 0, len(views)*2)
	for i, v := range views {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "($%d, $%d)", i*2+1, i*2+2)
		args = append(args, v.ProductID, v.ViewedAt)
	}

	_, err := r.db.ExecContext(ctx, sb.String(), args...)
	return err
}
