package search

import "context"

type SearchService interface {
	Search(ctx context.Context, searchModel interface{}) (*SearchResult, error)
}
