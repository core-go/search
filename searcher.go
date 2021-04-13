package search

import "context"

type Searcher struct {
	search func(ctx context.Context, searchModel interface{}, results interface{}, pageIndex int64, pageSize int64, options...int64) (int64, error)
}

func NewSearcher(search func(context.Context, interface{}, interface{}, int64, int64, ...int64) (int64, error)) *Searcher {
	return &Searcher{search: search}
}

func (s *Searcher) Search(ctx context.Context, m interface{}, results interface{}, pageIndex int64, pageSize int64, options...int64) (int64, error) {
	return s.search(ctx, m, results, pageIndex, pageSize, options...)
}
