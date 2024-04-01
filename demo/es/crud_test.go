package es

import (
	"context"
	"encoding/json"
	elastic "github.com/elastic/go-elasticsearch/v8"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

type ElasticSearchSuite struct {
	suite.Suite
	// 官方的
	es *elastic.Client
	// 一个对官方的封装
	olivere *olivere.Client
}

// SetupSuite 初始化客户端
func (s *ElasticSearchSuite) SetupSuite() {
	es, err := elastic.NewClient(elastic.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	require.NoError(s.T(), err)
	s.es = es
	ol, err := olivere.NewClient(olivere.SetURL("http://localhost:9200"))
	require.NoError(s.T(), err)
	s.olivere = ol
}

// TestCreateIndex 创建索引
func (s *ElasticSearchSuite) TestCreateIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	// 创建索引的请求
	def := `
	{  
	  "settings": {  
		"number_of_shards": 3,  
		"number_of_replicas": 2  
	  },  
	  "mappings": {  
		"properties": {
		  "email": {  
			"type": "text"  
		  },  
		  "phone": {  
			"type": "keyword"  
		  },  
		  "birthday": {  
			"type": "date"  
		  }
		}  
	  }  
	}
	`
	resp, err := s.es.Indices.Create("user_idx_go",
		s.es.Indices.Create.WithContext(ctx),
		s.es.Indices.Create.WithBody(strings.NewReader(def)))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)

	_, err = s.olivere.CreateIndex("user_idx_go_ol").Body(def).Do(ctx)
	require.NoError(s.T(), err)

}

// TestPutDoc 插入数据
func (s *ElasticSearchSuite) TestPutDoc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	data := `
{
	"email": "john@example.com",
	"phone": "1234567890",
	"birthday": "2000-01-01"
}
`
	_, err := s.es.Index("user_idx_go",
		strings.NewReader(data),
		s.es.Index.WithContext(ctx))
	require.NoError(s.T(), err)

	_, err = s.olivere.Index().Index("user_idx_go_ol").BodyString(data).Do(ctx)
	require.NoError(s.T(), err)
}

// TestGetDoc 搜索数据
func (s *ElasticSearchSuite) TestGetDoc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	query := `
{
	"query": {
		"range": {
			"birthday": {
				"gte": "1990-01_01"
			}
		}
	}
}
`

	// 这里返回的是 http 响应，想要拿到具体数据就要从其中进行解析
	_, err := s.es.Search(s.es.Search.WithIndex("user_idx_go"),
		s.es.Search.WithContext(ctx),
		s.es.Search.WithBody(strings.NewReader(query)))
	require.NoError(s.T(), err)

	// 使用的关键就是如何在对应的业务场景下构造复杂的查询
	olQuery := olivere.NewMatchQuery("email", "john")
	resp, err := s.olivere.Search("user_idx_go_ol").Query(olQuery).Do(ctx)
	require.NoError(s.T(), err)

	// 真正的数据在 Hits 里
	for _, hit := range resp.Hits.Hits {
		var u User
		err = json.Unmarshal(hit.Source, &u)
		require.NoError(s.T(), err)
		s.T().Log(u)
	}
}

func TestElasticSearch(t *testing.T) {
	suite.Run(t, new(ElasticSearchSuite))
}

type User struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}
