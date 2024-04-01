package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

const CollectIndexName = "collect_index"

type CollectESDAO struct {
	client *elastic.Client
}

func (c *CollectESDAO) Search(ctx context.Context, uid int64, biz string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
	)
	resp, err := c.client.Search(CollectIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var b BizInteractive
		err = json.Unmarshal(hit.Source, &b)
		if err != nil {
			return nil, err
		}
		res = append(res, b.BizId)
	}
	return res, nil
}

type BizInteractive struct {
	Id    int64  `json:"id"`
	Uid   int64  `json:"uid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
}
