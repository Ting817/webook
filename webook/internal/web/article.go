package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	l       logger.LoggerV1
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		biz:     "article",
		intrSvc: intrSvc,
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", a.Edit)
	g.POST("/publish", a.Publish)
	g.POST("/withdraw", a.Withdraw) // 仅自己可见
	g.POST("/list", ginx.WrapReqAndToken[Page, jwt.UserClaims](a.List))
	g.GET("/detail/:id", ginx.WrapToken[jwt.UserClaims](a.Detail))

	pub := g.Group("/pub")
	//pub.GET("/pub", a.PubList)
	pub.GET("/:id", ginx.WrapToken(a.PubDetail))
	pub.POST("/like", ginx.WrapReqAndToken[LikeReq, jwt.UserClaims](a.Like))
	pub.POST("/collect", ginx.WrapReqAndToken[CollectReq](a.Collect))
}

func (a *ArticleHandler) Edit(c *gin.Context) {
	var req ArticleReq
	if err := c.Bind(&req); err != nil {
		return
	}
	claim := c.MustGet("claims")
	claims, ok := claim.(*jwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	// 检测输入，跳过这一步
	// 调用 svc 的代码
	id, err := a.svc.Save(c, req.toDomain(claims.Uid))
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	c.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (a *ArticleHandler) Publish(c *gin.Context) {
	var req ArticleReq
	if err := c.Bind(&req); err != nil {
		return
	}
	claim := c.MustGet("claims")
	claims, ok := claim.(*jwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}

	id, err := a.svc.Publish(c, req.toDomain(claims.Uid))
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	c.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (a *ArticleHandler) Withdraw(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		a.l.Error("反序列化请求失败", logger.Error(err))
		return
	}

	usr, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}

	if err := a.svc.Withdraw(ctx, usr.Uid, req.Id); err != nil {
		a.l.Error("设置为仅自己可见失败", logger.Error(err), logger.Field{Key: "id", Value: req.Id})
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (a *ArticleHandler) List(ctx *gin.Context, req Page, uc jwt.UserClaims) (ginx.Result, error) {
	res, err := a.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "system error"}, err
	}
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Status:   src.Status.ToUint8(),
				// 这个列表请求，不需要返回内容
				//Content: src.Content,
				// 这个是创作者看自己的文章列表，也不需要这个字段
				//Author: src.Author
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (a *ArticleHandler) Detail(ctx *gin.Context, usr jwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return Result{Code: 4, Msg: "参数错误"}, err
	}

	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		return Result{Code: 5, Msg: "system error"}, err
	}

	if art.Author.Id != usr.Uid {
		return Result{Code: 4, Msg: "输入有误"}, fmt.Errorf("非法访问文章， 创作者 ID 不匹配 %w", err)
	}

	return Result{
		Data: ArticleVO{
			Id:       art.Id,
			Title:    art.Title,
			Abstract: art.Abstract(),
			Status:   art.Status.ToUint8(),
			Ctime:    art.Ctime.Format(time.DateTime),
			Utime:    art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context, uc ginx.UserClaims) (Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return Result{
			Code: 4,
			Msg:  "参数错误",
		}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idstr, err)
	}

	// 使用 error group 来同时查询数据
	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)
	eg.Go(func() error {
		var er error
		art, er = a.svc.GetPublishedById(ctx, id)
		return er
	})

	eg.Go(func() error {
		var er error
		intr, er = a.intrSvc.Get(ctx, a.biz, id, uc.Uid)
		return er
	})

	err = eg.Wait()

	if err != nil {
		return Result{
			Code: 5,
			Msg:  "系统错误",
		}, fmt.Errorf("获取文章信息失败 %w", err)
	}

	// 直接异步操作，在确定我们获取到了数据之后再来操作
	go func() {
		err = a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
		if err != nil {
			a.l.Error("增加文章阅读数失败", logger.Error(err))
		}
	}()

	return Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author:     art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
		},
	}, nil
}

func (a *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = a.intrSvc.Like(ctx, a.biz, req.Id, uc.Uid)
	} else {
		err = a.intrSvc.CancelLike(ctx, a.biz, req.Id, uc.Uid)
	}

	if err != nil {
		return Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return Result{Msg: "OK"}, nil
}

func (a *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc jwt.UserClaims) (Result, error) {
	err := a.intrSvc.Collect(ctx, a.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		return Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return Result{Msg: "OK"}, nil
}
