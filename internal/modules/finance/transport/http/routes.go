package financehttp

import "github.com/gin-gonic/gin"

func RegisterAdminRoutes(admin gin.IRoutes, handler *AdminHandler) {
	admin.GET("/finance/report", handler.Report)
	admin.GET("/finance/cash-flows", handler.ListCashFlows)
	admin.GET("/finance/cash-flow-report", handler.CashFlowReport)
	admin.GET("/finance/cash-flow-options", handler.CashFlowOptions)
	admin.POST("/finance/cash-flow-sync", handler.SyncZovoFinance)
	admin.POST("/finance/cash-flows", handler.CreateCashFlow)
	admin.PUT("/finance/cash-flows/:id", handler.UpdateCashFlow)
	admin.DELETE("/finance/cash-flows/:id", handler.DeleteCashFlow)
	admin.GET("/finance/receipts", handler.ListReceipts)
	admin.POST("/finance/receipts", handler.CreateReceipt)
	admin.PUT("/finance/receipts/:id", handler.UpdateReceipt)
	admin.DELETE("/finance/receipts/:id", handler.DeleteReceipt)
	admin.GET("/finance/cdk-deliveries", handler.ListDeliveries)
	admin.POST("/finance/cdk-deliveries", handler.CreateDelivery)
	admin.DELETE("/finance/cdk-deliveries/:id", handler.DeleteDelivery)
}
