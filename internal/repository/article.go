package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
	intrv2 "webook/api/proto/gen/intr/v2"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)

	// assignment 11
	GetIntr(ctx context.Context, biz string, id int64, uid int64) (*intrv2.Interactive, error)
	LikeIntr(ctx context.Context, biz string, id int64, uid int64) error
	CancelLikeIntr(ctx context.Context, biz string, id int64, uid int64) error
	CollectIntr(ctx context.Context, biz string, id int64, cid int64, uid int64) error
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	// 为什么这里用 UserRepository 而不是 UserDAO？
	// 因为你会绕开 repository
	// 而 repository 的核心职责是完成领域对象的构建，同时会有一些缓存机制我们不希望绕开
	userRepo UserRepository

	// 在 repository 层面同步数据（发表）
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO

	// SyncV2 要开启事务，肯定需要 db
	// 虽然能够通过事务实现制作库和线上库的数据同步，但是存在缺陷：
	// 没有面向接口原则，直接引入 gorm.DB 的依赖
	// 跨层依赖
	db *gorm.DB

	// 在 Repository 聚合微服务 Interactive
	intrRepo intrv2.InteractiveRepositoryClient
}

func (c *CachedArticleRepository) LikeIntr(ctx context.Context, biz string, id int64, uid int64) error {
	_, err := c.intrRepo.IncrLike(ctx, &intrv2.IncrLikeRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	return err
}

func (c *CachedArticleRepository) CancelLikeIntr(ctx context.Context, biz string, id int64, uid int64) error {
	_, err := c.intrRepo.DecrLike(ctx, &intrv2.DecrLikeRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	return err
}

func (c *CachedArticleRepository) CollectIntr(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	_, err := c.intrRepo.AddCollectionItem(ctx, &intrv2.AddCollectionItemRequest{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	return err
}

func (c *CachedArticleRepository) GetIntr(ctx context.Context, biz string, id int64, uid int64) (*intrv2.Interactive, error) {
	var intr *intrv2.Interactive
	// 拿阅读数
	respIntr, err := c.intrRepo.Get(ctx, &intrv2.GetRequest{
		Biz:   biz,
		BizId: id,
	})
	if err != nil {
		return nil, err
	}
	intr = respIntr.Intr

	// 拿是否点赞
	respLiked, err := c.intrRepo.Liked(ctx, &intrv2.LikedRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	intr.Liked = respLiked.Liked

	// 拿是否收藏
	respCollected, err := c.intrRepo.Collected(ctx, &intrv2.CollectedRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	if err != nil {
		return nil, err
	}
	intr.Collected = respCollected.Collected

	return intr, nil
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](arts, func(idx int, src dao.PublishedArticle) domain.Article {
		return c.toDomain(dao.Article(src))
	}), nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}

	// 现在要去查询对应的 User 信息，拿到创作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindUserInfoById(ctx, res.Author.Id)
	if err != nil {
		return domain.Article{}, err
		//return res, nil
	}
	res.Author.Name = author.NickName

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()

	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	go func() {
		er := c.cache.Set(ctx, c.toDomain(art))
		if er != nil {
			// 记录日志
		}
	}()
	return c.toDomain(art), nil
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 接入对第一页的缓存
	if offset == 0 && limit == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 记录日志
			// 区分缓存是否命中
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	var doArts []domain.Article
	for _, art := range arts {
		doArts = append(doArts, c.toDomain(art))
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败的处理思路，有可能是 redis 出问题了需要人工干预，也有可能只是网络抖动，忽略即可，因此这里考虑引入日志
			err = c.cache.SetFirstPage(ctx, uid, doArts)
			if err != nil {
				// 记录日志，监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, doArts)
	}()

	return doArts, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 记录日志
		}
	}
	return err

}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	// 缓存方案: 在发布的时候进行缓存，考虑到作者有粉丝，当发布时很可能被访问的情况
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		user, er := c.userRepo.FindUserInfoById(ctx, art.Author.Id)
		if er != nil {
			// 记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.NickName,
		}
		// SetPub 这里可以灵活设置缓存的过期时间
		// 当是大 V，粉丝多，应该设置更大的过期时间
		// 粉丝少，应该设置更小的过期时间
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()

	return id, err
}

// SyncV2 事务实现，前提是 repository 层面知道：底层数据存储用的是关系型数据库；同时制作库和线上库是一个数据库的两张表
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	authorDAO := dao.NewArticleGORMAuthorDAO(tx)
	readerDAO := dao.NewArticleGORMReaderDAO(tx)

	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}

	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		return 0, err
	}

	tx.Commit()
	return id, nil

}

// SyncV1 非事务实现
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = c.authorDAO.Update(ctx, artn)
	} else {
		id, err = c.authorDAO.Create(ctx, artn)
	}

	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err

}

func NewCachedArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, userRepo UserRepository, intrRepo intrv2.InteractiveRepositoryClient) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		intrRepo: intrRepo,
	}
}

func NewCachedArticleRepositoryV1(authorDAO dao.ArticleAuthorDAO, readerDAO dao.ArticleReaderDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		readerDAO: readerDAO,
		authorDAO: authorDAO,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

// preCache 业务相关的缓存预加载，预测用户大概率访问第一条，所以加载列表的时候，就将第一条的内容预加载到缓存中
func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 但是不缓存大文档（不是说所有情况都不缓存大对象，而是要根据情景权衡性能和内存开销）
	const size = 102 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录日志
		}
	}
}
