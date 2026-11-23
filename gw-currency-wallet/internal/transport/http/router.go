package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/auth"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(h *Handler, authSvc *auth.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// публичные
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)

		// защищённые
		protected := api.Group("")
		protected.Use(JWTMiddleware(authSvc))
		{
			protected.GET("/balance", h.GetBalance)
			protected.POST("/wallet/deposit", h.Deposit)
			protected.POST("/wallet/withdraw", h.Withdraw)
			protected.GET("/exchange/rates", h.GetRates)
			protected.POST("/exchange", h.Exchange)
		}
	}

	return r
}
