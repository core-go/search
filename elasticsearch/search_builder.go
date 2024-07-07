package elasticsearch

import (
	"context"
	"fmt"
	"reflect"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchBuilder[T any, F any] struct {
	Client      *elasticsearch.Client
	Index       []string
	BuildQuery  func(F) map[string]interface{}
	GetSort     func(interface{}) string
	ModelType   reflect.Type
	idJson      string
	versionJson string
	Map         func(*T)
}

func NewSearchBuilder[T any, F any](client *elasticsearch.Client, index []string, buildQuery func(F) map[string]interface{}, getSort func(m interface{}) string, opts ...func(*T)) *SearchBuilder[T, F] {
	return NewSearchBuilderWithVersion[T, F](client, index, buildQuery, getSort, "", opts...)
}
func NewSearchBuilderWithVersion[T any, F any](client *elasticsearch.Client, index []string, buildQuery func(F) map[string]interface{}, getSort func(m interface{}) string, versionJson string, opts ...func(*T)) *SearchBuilder[T, F] {
	var t T
	modelType := reflect.TypeOf(t)
	if modelType.Kind() != reflect.Struct {
		panic("T must be a struct")
	}
	idIndex, _, idJson := FindIdField(modelType)
	if idIndex < 0 {
		panic(fmt.Sprintf("%s struct requires id field which bson name is '_id'", modelType.Name()))
	}
	var mp func(*T)
	if len(opts) > 0 && opts[0] != nil {
		mp = opts[0]
	}
	return &SearchBuilder[T, F]{Client: client, Index: index, BuildQuery: buildQuery, GetSort: getSort, ModelType: modelType, idJson: idJson, versionJson: versionJson, Map: mp}
}
func (b *SearchBuilder[T, F]) Search(ctx context.Context, filter F, limit int64, offset int64) ([]T, int64, error) {
	query := b.BuildQuery(filter)
	s := b.GetSort(filter)
	sort := BuildSort(s, b.ModelType)
	var objs []T
	total, err := BuildSearchResult(ctx, b.Client, b.Index, objs, b.idJson, query, sort, limit, offset, b.versionJson)
	if b.Map != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			b.Map(&objs[i])
		}
	}
	return objs, total, err
}
