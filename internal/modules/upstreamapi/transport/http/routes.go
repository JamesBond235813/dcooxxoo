package upstreamhttp

import "github.com/gin-gonic/gin"

func RegisterAuthenticatedRoutes(upstream gin.IRoutes, handler *Handler) {
	if upstream == nil || handler == nil {
		panic("upstream authenticated routes: required dependency is nil")
	}
	upstream.POST("/ping", handler.Ping)
	upstream.GET("/categories", handler.ListCategories)
	upstream.GET("/products", handler.ListProducts)
	upstream.GET("/products/:id", handler.GetProduct)
	upstream.POST("/orders", handler.CreateOrder)
	upstream.GET("/orders/:id", handler.GetOrder)
	upstream.POST("/orders/:id/cancel", handler.CancelOrder)
	upstream.GET("/agent/products", handler.ListAgentProducts)
	upstream.GET("/agent/products/:id", handler.GetAgentProduct)
	upstream.POST("/agent/orders", handler.CreateAgentOrder)
	upstream.GET("/agent/orders/:id", handler.GetAgentOrder)
	upstream.POST("/agent/orders/:id/session/preflight", handler.PreflightAgentSession)
	upstream.POST("/agent/orders/:id/session", handler.SubmitAgentSession)
}
