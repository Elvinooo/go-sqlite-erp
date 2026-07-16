package router

import (
	authapp "erp/internal/application/auth"
	businessapp "erp/internal/application/business"
	customerapp "erp/internal/application/customer"
	dashboardapp "erp/internal/application/dashboard"
	supplierapp "erp/internal/application/supplier"
	systemapp "erp/internal/application/system"
	"erp/internal/infrastructure/security"
	"erp/internal/interfaces/http/controller"
	"erp/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, jwt *security.JWTManager, authService *authapp.Service, systemService *systemapp.Service, customerService *customerapp.Service, supplierService *supplierapp.Service, dashboardService *dashboardapp.Service, businessService *businessapp.Service) {
	authController := controller.NewAuthController(authService)
	auditController := controller.NewAuditController(authService)
	userController := controller.NewUserController(authService)
	roleController := controller.NewRoleController(authService)
	permissionController := controller.NewPermissionController(authService)
	menuController := controller.NewMenuController(authService)
	settingController := controller.NewSettingController(systemService)
	customerController := controller.NewCustomerController(customerService)
	supplierController := controller.NewSupplierController(supplierService)
	dashboardController := controller.NewDashboardController(dashboardService)
	businessController := controller.NewBusinessController(businessService)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ERP Pro API service is running",
			"data": gin.H{
				"frontend": "http://localhost:5173/login",
				"apiBase":  "http://127.0.0.1:18080/api/v1",
				"health":   "http://127.0.0.1:18080/healthz",
			},
		})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
	})
	r.Static("/uploads", "data/uploads")
	r.GET("/print/:module/:id", businessController.Print)
	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", authController.Login)
		api.POST("/auth/refresh", authController.Refresh)
	}
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jwt, authService), middleware.DataScope(authService))
	{
		protected.GET("/auth/me", authController.Me)
		protected.POST("/auth/logout", authController.Logout)
		protected.PUT("/auth/password", authController.ChangePassword)
		protected.PUT("/auth/profile", authController.UpdateProfile)
		protected.GET("/dashboard/boss", middleware.RBAC(authService, "dashboard.boss.view"), dashboardController.Boss)
		protected.GET("/dashboard/price-profit", middleware.RBAC(authService, "dashboard.boss.view"), dashboardController.PriceProfit)
		protected.POST("/dashboard/sign-in", middleware.RBAC(authService, "dashboard.boss.view"), dashboardController.SignIn)
		protected.GET("/dashboard/sign-ins", dashboardController.SignInHistory)
		protected.GET("/business/modules", businessController.Modules)
		protected.GET("/business/:module/meta", businessController.Meta)
		registerBusinessModule(protected, authService, "products", "product.manage", businessController)
		registerBusinessModule(protected, authService, "sales", "sales.manage", businessController)
		registerBusinessModule(protected, authService, "purchase", "purchase.manage", businessController)
		registerBusinessModule(protected, authService, "inventory", "inventory.manage", businessController)
		registerBusinessModule(protected, authService, "inventory-movements", "inventory.manage", businessController)
		registerBusinessModule(protected, authService, "repair", "repair.manage", businessController)
		registerBusinessModule(protected, authService, "project", "project.manage", businessController)
		registerBusinessModule(protected, authService, "finance", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "finance-accounts", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "receivables", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "customer-statements", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "payables", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "profit-report", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "inventory-asset-report", "finance.manage", businessController)
		registerBusinessModule(protected, authService, "document-delete-records", "system.audit.view", businessController)

		users := protected.Group("/users", middleware.RBAC(authService, "auth.user.manage"))
		users.GET("", userController.List)
		users.POST("", userController.Create)
		users.PUT("/:id", userController.Update)
		users.PUT("/:id/password", middleware.RBAC(authService, "auth.user.reset_password"), userController.ResetPassword)
		users.DELETE("/:id", userController.Delete)

		roles := protected.Group("/roles", middleware.RBAC(authService, "auth.role.manage"))
		roles.GET("", roleController.List)
		roles.POST("", roleController.Create)
		roles.PUT("/:id", roleController.Update)
		roles.DELETE("/:id", roleController.Delete)

		permissions := protected.Group("/permissions", middleware.RBAC(authService, "auth.permission.manage"))
		permissions.GET("", permissionController.List)
		permissions.POST("", permissionController.Create)
		permissions.PUT("/:id", permissionController.Update)
		permissions.DELETE("/:id", permissionController.Delete)

		menus := protected.Group("/menus", middleware.RBAC(authService, "auth.menu.manage"))
		menus.GET("", menuController.List)
		menus.GET("/tree", menuController.Tree)
		menus.POST("", menuController.Create)
		menus.PUT("/:id", menuController.Update)
		menus.DELETE("/:id", menuController.Delete)

		settings := protected.Group("/settings", middleware.RBAC(authService, "system.setting.manage"))
		settings.GET("", settingController.List)
		settings.GET("/merchant-info", settingController.MerchantInfo)
		settings.PUT("/merchant-info", settingController.SaveMerchantInfo)
		settings.POST("", settingController.Create)
		settings.POST("/restore-test-data", settingController.RestoreTestData)
		settings.PUT("/:id", settingController.Update)
		settings.DELETE("/:id", settingController.Delete)

		customers := protected.Group("/customers", middleware.RBAC(authService, "customer.manage"))
		customers.GET("", customerController.List)
		customers.POST("", customerController.Create)
		customers.POST("/import", customerController.ImportExcel)
		customers.GET("/export", customerController.ExportExcel)
		customers.GET("/:id", customerController.Get)
		customers.PUT("/:id", customerController.Update)
		customers.DELETE("/:id", customerController.Delete)
		customers.GET("/:id/debt", customerController.Debt)
		customers.GET("/:id/orders", customerController.OrderHistory)

		suppliers := protected.Group("/suppliers", middleware.RBAC(authService, "supplier.manage"))
		suppliers.GET("", supplierController.List)
		suppliers.POST("", supplierController.Create)
		suppliers.GET("/:id", supplierController.Get)
		suppliers.PUT("/:id", supplierController.Update)
		suppliers.DELETE("/:id", supplierController.Delete)

		audit := protected.Group("/audit", middleware.RBAC(authService, "system.audit.view"))
		audit.GET("/login-logs", auditController.LoginLogs)
		audit.GET("/operation-logs", auditController.OperationLogs)
	}
}

func registerBusinessModule(parent *gin.RouterGroup, authService *authapp.Service, module string, permission string, controller *controller.BusinessController) {
	group := parent.Group("/"+module, middleware.RBAC(authService, permission))
	group.Use(func(c *gin.Context) {
		c.Params = append(c.Params, gin.Param{Key: "module", Value: module})
		c.Next()
	})
	group.GET("", controller.List)
	group.POST("", controller.Create)
	group.GET("/:id", controller.Get)
	group.GET("/:id/photos", controller.ListPhotos)
	group.POST("/:id/photos", controller.UploadPhoto)
	group.PUT("/:id", controller.Update)
	group.DELETE("/:id", middleware.RBAC(authService, "document_delete"), controller.Delete)
	group.POST("/actions", controller.Action)
}
