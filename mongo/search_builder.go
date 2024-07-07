package mongo

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SearchBuilder[T any, F any] struct {
	Collection *mongo.Collection
	BuildQuery func(m F) (bson.D, bson.M)
	GetSort    func(m interface{}) string
	BuildSort  func(s string, modelType reflect.Type) bson.D
	Map        func(*T)
}

func NewSearchQueryWithSort[T any, F any](db *mongo.Database, collectionName string, buildQuery func(F) (bson.D, bson.M), getSort func(interface{}) string, buildSort func(string, reflect.Type) bson.D, options ...func(*T)) *SearchBuilder[T, F] {
	return NewSearchBuilderWithSort[T, F](db, collectionName, buildQuery, getSort, buildSort, options...)
}
func NewSearchQuery[T any, F any](db *mongo.Database, collectionName string, buildQuery func(F) (bson.D, bson.M), getSort func(interface{}) string, options ...func(*T)) *SearchBuilder[T, F] {
	return NewSearchBuilderWithSort[T, F](db, collectionName, buildQuery, getSort, BuildSort, options...)
}
func NewSearchBuilderWithSort[T any, F any](db *mongo.Database, collectionName string, buildQuery func(F) (bson.D, bson.M), getSort func(interface{}) string, buildSort func(string, reflect.Type) bson.D, options ...func(*T)) *SearchBuilder[T, F] {
	var mp func(*T)
	if len(options) > 0 && options[0] != nil {
		mp = options[0]
	}
	collection := db.Collection(collectionName)
	builder := &SearchBuilder[T, F]{Collection: collection, BuildQuery: buildQuery, GetSort: getSort, BuildSort: buildSort, Map: mp}
	return builder
}
func NewSearchBuilder[T any, F any](db *mongo.Database, collectionName string, buildQuery func(F) (bson.D, bson.M), getSort func(interface{}) string, options ...func(*T)) *SearchBuilder[T, F] {
	return NewSearchBuilderWithSort[T, F](db, collectionName, buildQuery, getSort, BuildSort, options...)
}

func (b *SearchBuilder[T, F]) Search(ctx context.Context, m F, limit int64, skip int64) ([]T, int64, error) {
	var objs []T
	query, fields := b.BuildQuery(m)

	var sort = bson.D{}
	s := b.GetSort(m)
	modelType := reflect.TypeOf(&objs).Elem().Elem()
	sort = b.BuildSort(s, modelType)
	if skip < 0 {
		skip = 0
	}
	total, err := BuildSearchResult(ctx, b.Collection, &objs, query, fields, sort, limit, skip)
	if b.Map != nil {
		l := len(objs)
		for i := 0; i < l; i++ {
			b.Map(&objs[i])
		}
	}
	return objs, total, err
}
