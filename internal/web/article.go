package web

import (
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/jwt"
	"webook/pkg/logger"
)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc intrv1.InteractiveServiceClient
	l        logger.LoggerV1
	biz      string
}

func NewArticleHandler(l logger.LoggerV1, svc service.ArticleService, interSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	//func NewArticleHandler(l logger.LoggerV1, svc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		l:        l,
		svc:      svc,
		interSvc: interSvc,
		biz:      "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	// 这里演示了基于集成测试的TDD
	g.POST("/edit", h.Edit)
	// 这里演示了基于单元测试的TDD
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// 创作者接口
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", h.List)

	// 读者接口
	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true 点赞 false 不点赞
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)

	// assignment week9
	pub.POST("")
}

// Edit 暂时约定，接收 Article 输入，返回文章 ID
func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 从登录态拿到用户信息
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.UserId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.l.Error("保存文章数据失败",
			logger.Int64("uid", uc.UserId),
			logger.Error(err))
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})

}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 从登录态拿到用户信息
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.UserId,
		},
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表文章数据失败",
			logger.Int64("uid", uc.UserId),
			logger.Error(err))
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 从登录态拿到用户信息
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.svc.Withdraw(ctx, uc.UserId, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("撤回文章失败",
			logger.Int64("uid", uc.UserId),
			logger.Int64("aid", req.Id),
			logger.Error(err))
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.UserId, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.UserId),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				//Content:  src.Content,
				AuthorId: src.Author.Id,
				Status:   src.Status.ToUint8(),
				Ctime:    src.Ctime.Format(time.DateTime),
				Utime:    src.Utime.Format(time.DateTime),
			}
		}),
	})

}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	// 拿到文章 ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Warn("查询文章失败, id 格式不对",
			logger.Error(err),
			logger.String("id", idStr),
		)
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败",
			logger.Error(err),
			logger.Int64("id", id),
		)
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.UserId {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("非法查询：作者不匹配",
			logger.Error(err),
			logger.Int64("aid", art.Author.Id),
			logger.Int64("uid", uc.UserId),
		)
		return
	}

	vo := ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),
		Content: art.Content,
		//AuthorId: art.Author.Id,
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}

	ctx.JSON(http.StatusOK, Result{
		Data: vo,
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	// 拿到文章 ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Warn("查询文章失败, id 格式不对",
			logger.Error(err),
			logger.String("id", idStr),
		)
		return
	}

	//go func() {
	//	newCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//	defer cancel()
	//	er := h.interSvc.IncrReadCnt(newCtx, h.biz, id)
	//	if er != nil {
	//		// 记录日志
	//		h.l.Error("更新阅读数失败",
	//			logger.Int64("aid", id),
	//			logger.Error(err))
	//	}
	//}()

	var (
		eg       errgroup.Group
		art      domain.Article
		intrResp *intrv1.GetResponse

		// assignment 11
		//intrResp2 *intrv2.Interactive
	)

	uc := ctx.MustGet("user").(jwt.UserClaims)
	// 获取文章，这里就会触发 kafka 发送消息
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id, uc.UserId)
		return er
	})

	// 获取阅读数/点赞/收藏 （已微服务化）
	eg.Go(func() error {
		var er error
		intrResp, er = h.interSvc.Get(ctx, &intrv1.GetRequest{
			Biz: h.biz, BizId: id, Uid: uc.UserId,
		})
		// assignment 11
		//intrResp2, er = h.svc.GetIntr(ctx, h.biz, id, uc.UserId)
		return er
	})

	err = eg.Wait()

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败，系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uc.UserId),
			logger.Error(err),
		)
		return
	}

	intr := intrResp.Intr
	vo := ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),
		Content: art.Content,
		//AuthorId: art.Author.Id,
		AuthorName: art.Author.Name,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,

		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}

	// assignment 11
	//vo := ArticleVo{
	//	Id:    art.Id,
	//	Title: art.Title,
	//	//Abstract: art.Abstract(),
	//	Content: art.Content,
	//	//AuthorId: art.Author.Id,
	//	AuthorName: art.Author.Name,
	//	ReadCnt:    intrResp2.ReadCnt,
	//	LikeCnt:    intrResp2.LikeCnt,
	//	CollectCnt: intrResp2.CollectCnt,
	//	Liked:      intrResp2.Liked,
	//	Collected:  intrResp2.Collected,
	//
	//	Status: art.Status.ToUint8(),
	//	Ctime:  art.Ctime.Format(time.DateTime),
	//	Utime:  art.Utime.Format(time.DateTime),
	//}
	ctx.JSON(http.StatusOK, Result{
		Data: vo,
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	var err error
	if req.Like {
		_, err = h.interSvc.Like(ctx, &intrv1.LikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.UserId,
		})

		//err = h.svc.LikeIntr(ctx, h.biz, req.Id, uc.UserId)
	} else {
		_, err = h.interSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.UserId,
		})
		//err = h.svc.CancelLikeIntr(ctx, h.biz, req.Id, uc.UserId)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("uid", uc.UserId),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	_, err := h.interSvc.Collect(ctx, &intrv1.CollectRequest{
		Biz: h.biz, BizId: req.Id, Cid: req.Cid, Uid: uc.UserId,
	})
	//err := h.svc.CollectIntr(ctx, h.biz, req.Id, req.Cid, uc.UserId)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", uc.UserId),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}
