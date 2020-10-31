package search

type QueryBuilder interface {
	BuildQuery(sm interface{}) (string, []interface{})
}
