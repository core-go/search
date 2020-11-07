package search

import "context"

type SearchResultBuilder interface {
	BuildSearchResult(ctx context.Context, m interface{}) (interface{}, int64, error)
}
