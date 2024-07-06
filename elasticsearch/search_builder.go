package elasticsearch

import (
	"context"
	"fmt"
	"reflect"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchBuilder struct {
	Client      *elasticsearch.Client
	Index       []string
	BuildQuery  func(searchModel interface{}) map[string]interface{}
	GetSort     func(m interface{}) string
	ModelType   reflect.Type
	idJson      string
	versionJson string
}

func NewSearchBuilder(client *elasticsearch.Client, index []string, modelType reflect.Type, buildQuery func(interface{}) map[string]interface{}, getSort func(m interface{}) string) *SearchBuilder {
	return NewSearchBuilderWithVersion(client, index, modelType, buildQuery, getSort, "")
}
func NewSearchBuilderWithVersion(client *elasticsearch.Client, index []string, modelType reflect.Type, buildQuery func(interface{}) map[string]interface{}, getSort func(m interface{}) string, versionJson string) *SearchBuilder {
	idIndex, _, idJson := FindIdField(modelType)
	if idIndex < 0 {
		panic(fmt.Sprintf("%s struct requires id field which bson name is '_id'", modelType.Name()))
	}
	return &SearchBuilder{Client: client, Index: index, BuildQuery: buildQuery, GetSort: getSort, ModelType: modelType, idJson: idJson, versionJson: versionJson}
}
func (b *SearchBuilder) Search(ctx context.Context, filter interface{}, results interface{}, limit int64, offset int64) (int64, error) {
	query := b.BuildQuery(filter)
	s := b.GetSort(filter)
	sort := BuildSort(s, b.ModelType)
	total, err := BuildSearchResult(ctx, b.Client, b.Index, results, b.idJson, query, sort, limit, offset, b.versionJson)
	return total, err
}
