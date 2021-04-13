package search

import "context"

type SearchService interface {
	Search(ctx context.Context, searchModel interface{}, results interface{}, pageIndex int64, pageSize int64, options...int64) (int64, error)
}
