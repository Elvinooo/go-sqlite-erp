package bootstrap

import (
	"fmt"

	authapp "erp/internal/application/auth"
	businessapp "erp/internal/application/business"
	customerapp "erp/internal/application/customer"
	dashboardapp "erp/internal/application/dashboard"
	supplierapp "erp/internal/application/supplier"
	systemapp "erp/internal/application/system"
	"erp/internal/config"
	"erp/internal/infrastructure/database"
	"erp/internal/infrastructure/logger"
	gormrepo "erp/internal/infrastructure/persistence/gorm"
	"erp/internal/infrastructure/security"
	"erp/internal/interfaces/http/middleware"
	"erp/internal/interfaces/http/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	db     *gorm.DB
	log    *zap.Logger
	engine *gin.Engine
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	log, err := logger.New(cfg.Log)
	if err != nil {
		return nil, err
	}
	db, err := database.Open(cfg.Database)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db, cfg.Admin); err != nil {
		return nil, err
	}

	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	engine.MaxMultipartMemory = 8 << 20
	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	engine.Use(middleware.Recovery(log), middleware.CORS(), middleware.RequestLogger(log), middleware.DemoMode(cfg.App.DemoMode), middleware.OperationAudit(db), middleware.ErrorHandler(log))

	jwtManager := security.NewJWTManager(cfg.JWT)
	userRepo := gormrepo.NewUserRepository(db)
	roleRepo := gormrepo.NewRoleRepository(db)
	permissionRepo := gormrepo.NewPermissionRepository(db)
	menuRepo := gormrepo.NewMenuRepository(db)
	auditRepo := gormrepo.NewAuditRepository(db)
	settingRepo := gormrepo.NewSettingRepository(db)
	customerRepo := gormrepo.NewCustomerRepository(db)
	supplierRepo := gormrepo.NewSupplierRepository(db)
	dashboardRepo := gormrepo.NewBossDashboardRepository(db)
	moduleRepo := gormrepo.NewModuleRepository(db)

	authService := authapp.NewService(userRepo, roleRepo, permissionRepo, menuRepo, auditRepo, jwtManager)
	systemService := systemapp.NewService(settingRepo)
	customerService := customerapp.NewService(customerRepo)
	supplierService := supplierapp.NewService(supplierRepo)
	dashboardService := dashboardapp.NewService(dashboardRepo)
	businessService := businessapp.NewService(moduleRepo)
	router.Register(engine, jwtManager, authService, systemService, customerService, supplierService, dashboardService, businessService)

	return &App{cfg: cfg, db: db, log: log, engine: engine}, nil
}

func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%s", a.cfg.Server.Host, a.cfg.Server.Port)
	a.log.Info("erp backend started", zap.String("addr", addr))
	return a.engine.Run(addr)
}
