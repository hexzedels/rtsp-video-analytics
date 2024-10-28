package job

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"streaming/api/internal/streaming/pb"
)

func (r *jobPrivateRouter) new(ctx *gin.Context) {
	var req pb.RunQuery

	if err := ctx.Bind(&req); err != nil {
		r.logger.Error("failed to bind run query", zap.Error(err))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	resp, err := r.orchestratorClient.Run(ctx, &req)
	if err != nil {
		r.logger.Error("orchestrator run request failed", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
