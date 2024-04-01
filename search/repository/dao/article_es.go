package dao

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

const ArticleIndexName = "article_index"
const TagIndexName = "tags_index"

type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type ArticleElasticDAO struct {
	client *elastic.Client
}

func NewArticleElasticDAO(client *elastic.Client) ArticleDAO {
	return &ArticleElasticDAO{client: client}
}

func (h *ArticleElasticDAO) Search(ctx context.Context, artIds []int64, colIds []int64, likeIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	// 2=> published
	status := elastic.NewTermQuery("status", 2)

	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)
	tag := elastic.NewTermsQuery("id", slice.Map(artIds, func(idx int, src int64) any {
		return src
	})).Boost(2)

	// assignment week19
	// 我们希望点赞比标签优先级更高，同时设置更高的相关性权重给收藏
	like := elastic.NewTermsQuery("id", slice.Map(likeIds, func(idx int, src int64) any {
		return src
	})).Boost(10)
	collect := elastic.NewTermsQuery("id", slice.Map(colIds, func(idx int, src int64) any {
		return src
	})).Boost(5)

	or := elastic.NewBoolQuery().Should(title, content, tag, collect, like)
	query := elastic.NewBoolQuery().Must(status, or)
	resp, err := h.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var art Article
		err = json.Unmarshal(hit.Source, &art)
		if err != nil {
			return nil, err
		}
		res = append(res, art)
	}
	return res, nil
}

func NewArticleRepository(client *elastic.Client) ArticleDAO {
	return &ArticleElasticDAO{
		client: client,
	}
}
func (h *ArticleElasticDAO) InputArticle(ctx context.Context, art Article) error {
	_, err := h.client.Index().Index(ArticleIndexName).
		Id(strconv.FormatInt(art.Id, 10)).
		BodyJson(art).Do(ctx)
	return err
}
