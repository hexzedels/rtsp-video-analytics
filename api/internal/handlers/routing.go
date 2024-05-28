package handlers

import (
	"streaming/api/internal/router"

	"github.com/gin-gonic/gin"
)

type jobPrivateRouter struct {
	*router.Private
}

func AttachToGroup(group *gin.RouterGroup, private *router.Private) {
	jobRouter := jobPrivateRouter{private}

	group.POST("/", jobRouter.new)
	group.GET("/:id", jobRouter.get)
	group.POST("/:id", jobRouter.post)
}
