package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CocaineCong/gin-mall/pkg/utils/ctl"
	"github.com/CocaineCong/gin-mall/pkg/utils/log"
	"github.com/CocaineCong/gin-mall/service"
	"github.com/CocaineCong/gin-mall/types"
)

func ImportSkillProductHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.SkillProductImportReq
		if err := ctx.ShouldBind(&req); err != nil {
			// 参数校验
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(ctx, err))
			return
		}

		file, _, _ := ctx.Request.FormFile("file")
		l := service.GetSkillProductSrv()
		resp, err := l.Import(ctx.Request.Context(), file)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}

func InitSkillProductHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.SkillProductImportReq
		if err := ctx.ShouldBind(&req); err != nil {
			// 参数校验
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(ctx, err))
			return
		}

		l := service.GetSkillProductSrv()
		resp, err := l.InitSkillGoods(ctx.Request.Context())
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}

func SkillProductHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req types.SkillProductServiceReq
		if err := ctx.ShouldBind(&req); err != nil {
			// 参数校验
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(ctx, err))
			return
		}

		l := service.GetSkillProductSrv()
		resp, err := l.SkillProduct(ctx.Request.Context(), &req)
		if err != nil {
			log.LogrusObj.Infoln(err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse(ctx, err))
			return
		}
		ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, resp))
		return
	}
}
