# Search Service for Golang
- Search Model
- Search Result With Config
- Search Service

## Installation
Please make sure to initialize a Go module before installing common-go/search:

```shell
go get -u github.com/common-go/search
```

Import:

```go
import "github.com/common-go/search"
```

## Details:
#### search_model.go
```go
package search

import "reflect"

type SearchModel struct {
	PageIndex     int64                    `mapstructure:"page_index" json:"pageIndex,omitempty" gorm:"column:pageindex" bson:"pageIndex,omitempty" dynamodbav:"pageIndex,omitempty" firestore:"pageIndex,omitempty"`
	PageSize      int64                    `mapstructure:"page_size" json:"pageSize,omitempty" gorm:"column:pagesize" bson:"pageSize,omitempty" dynamodbav:"pageSize,omitempty" firestore:"pageSize,omitempty"`
	FirstPageSize int64                    `mapstructure:"first_page_size" json:"firstPageSize,omitempty" gorm:"column:firstpagesize" bson:"firstPageSize,omitempty" dynamodbav:"firstPageSize,omitempty" firestore:"firstPageSize,omitempty"`
	Fields        []string                 `mapstructure:"fields" json:"fields,omitempty" gorm:"column:fields" bson:"fields,omitempty" dynamodbav:"fields,omitempty" firestore:"fields,omitempty"`
	Sort          string                   `mapstructure:"sort" json:"sort,omitempty" gorm:"column:sortfield" bson:"sort,omitempty" dynamodbav:"sort,omitempty" firestore:"sort,omitempty"`
	CurrentUserId string                   `mapstructure:"current_user_id" json:"currentUserId,omitempty" gorm:"column:currentuserid" bson:"currentUserId,omitempty" dynamodbav:"currentUserId,omitempty" firestore:"currentUserId,omitempty"`
	Keyword       string                   `mapstructure:"keyword" json:"keyword,omitempty" gorm:"column:keyword" bson:"keyword,omitempty" dynamodbav:"keyword,omitempty" firestore:"keyword,omitempty"`
	Excluding     map[string][]interface{} `mapstructure:"excluding" json:"excluding,omitempty" gorm:"column:excluding" bson:"excluding,omitempty" dynamodbav:"excluding,omitempty" firestore:"excluding,omitempty"`
	RefId         string                   `mapstructure:"refid" json:"refId,omitempty" gorm:"column:refid" bson:"refId,omitempty" dynamodbav:"refId,omitempty" firestore:"refId,omitempty"`
}
```
#### search_service.go
```go
package search

import "context"

type SearchService interface {
	Search(ctx context.Context, searchModel interface{}) (interface{}, int64, error)
}
```
