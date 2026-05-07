package middleware

import (
	"fayhub/internal/service"
	"fayhub/pkg/utils"

	"github.com/gin-gonic/gin"
)

func DataPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserIDFromContext(c)
		if !ok || userID == 0 {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		if utils.IsDataPermissionSkipped(ctx) {
			c.Next()
			return
		}

		permSvc := &service.DataPermissionService{}
		filter, err := permSvc.GetDataScope(ctx, userID)
		if err != nil || filter == nil {
			c.Next()
			return
		}

		ctx = utils.WithDataScopeFilter(ctx, filter)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
