package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/command"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/application/port/output"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/config"
	domainErrors "github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/domain-errors"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/repository"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/user"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/domain/valueobject"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/handler"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/http/middleware"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/persistence"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/persistence/postgres"
	"github.com/FranciscoHonorat/ordemflow/services/user-service/internal/infrastructure/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Get()
	log.Printf("Starting user-service with config: %+v", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := postgres.NewConnection(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established successfully.")

	uow := postgres.NewUnitOfWork(db)
	clock := persistence.NewRealClock()
	hasher := security.NewBcryptHasher(0)
	issuer := security.NewJWTIssuer(cfg.JWT.Secret, cfg.JWT.Expiry)
	userRepo := postgres.NewUserRepository(db, postgres.NewUserMapper())

	if cfg.BootstrapAdmin.Enabled() {
		if err := bootstrapAdmin(ctx, uow, clock, hasher, userRepo, cfg.BootstrapAdmin.Email, cfg.BootstrapAdmin.Password); err != nil {
			log.Fatalf("Failed to bootstrap admin user: %v", err)
		}
	}

	registerHandler := command.NewRegisterUserHandler(uow, clock, hasher)
	changeRoleHandler := command.NewChangeRoleHandler(uow, clock)
	activateHandler := command.NewActivateUserHandler(uow, clock)
	deactivateHandler := command.NewDeactivateUserHandler(uow, clock)
	loginHandler := command.NewLoginHandler(userRepo, hasher, issuer, clock)

	registerUC := &registerUserUseCaseAdapter{handler: registerHandler}
	changeRoleUC := &changeRoleUseCaseAdapter{handler: changeRoleHandler}
	activateUC := &activateUserUseCaseAdapter{handler: activateHandler}
	deactivateUC := &deactivateUserUseCaseAdapter{handler: deactivateHandler}
	loginUC := &loginUseCaseAdapter{handler: loginHandler}

	userQueries := postgres.NewUserQueries(db)

	authHandler := handler.NewAuthHandler(registerUC, loginUC)
	userHandler := handler.NewUserHandler(userQueries, changeRoleUC, activateUC, deactivateUC)
	healthHandler := handler.NewHealthHandler()

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", gin.WrapF(healthHandler.HealthCheck))

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	verifier := &tokenVerifierAdapter{issuer: issuer}

	authenticated := r.Group("/users")
	authenticated.Use(middleware.Auth(verifier))
	{
		authenticated.GET("/me", userHandler.GetMe)
	}

	admin := r.Group("/users")
	admin.Use(middleware.Auth(verifier), middleware.RequireRole(valueobject.RoleNameAdmin))
	{
		admin.GET("", userHandler.ListUsers)
		admin.GET("/:id", userHandler.GetUserByID)
		admin.PATCH("/:id/role", userHandler.ChangeRole)
		admin.POST("/:id/activate", userHandler.ActivateUser)
		admin.POST("/:id/deactivate", userHandler.DeactivateUser)
	}

	serverAddr := ":" + cfg.HTTP.Port
	if cfg.HTTP.Port == "" {
		serverAddr = ":8083"
	}

	log.Printf("Starting HTTP server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func bootstrapAdmin(
	ctx context.Context,
	uow output.UnitOfWork,
	clock output.Clock,
	hasher output.PasswordHasher,
	userRepo repository.UserRepository,
	email, password string,
) error {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return fmt.Errorf("invalid BOOTSTRAP_ADMIN_EMAIL: %w", err)
	}

	_, err = userRepo.FindByEmail(ctx, emailVO)
	if err == nil {
		log.Println("Bootstrap admin already exists, skipping.")
		return nil
	}
	if !errors.Is(err, domainErrors.ErrUserNotFound) {
		return fmt.Errorf("failed to check for existing bootstrap admin: %w", err)
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash bootstrap admin password: %w", err)
	}

	now := clock.Now()
	admin, err := user.NewUser(valueobject.NewUserIDMust(uuid.New()), emailVO, hash, valueobject.NewRoleMust(valueobject.RoleNameAdmin), now)
	if err != nil {
		return fmt.Errorf("failed to build bootstrap admin: %w", err)
	}

	err = uow.Do(ctx, func(store output.RepositoryProvider) error {
		if err := store.UserRepository().Save(ctx, admin); err != nil {
			return err
		}
		return store.OutboxRepository().SaveEvents(ctx, admin.PullEvents())
	})
	if err != nil {
		return fmt.Errorf("failed to save bootstrap admin: %w", err)
	}

	log.Println("Bootstrap admin created successfully.")
	return nil
}
