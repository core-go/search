package search

import "context"

type SearchBuilder interface {
	Search(ctx context.Context, m interface{}) (interface{}, int64, error)
}
