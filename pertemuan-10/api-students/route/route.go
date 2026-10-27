package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"api-students-v10/app/service"
	"api-students-v10/helper"
	"api-students-v10/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	Mongo             *mongo.Client
	JWT               *helper.JWTManager
	Permissions       *helper.PermissionSet
	StudentService    *service.StudentService
	AuthService       *service.AuthService
	CourseService     *service.CourseService
	SupervisorService *service.SupervisorService
	ReviewService     *service.ReviewService
	PortfolioService  *service.PortfolioService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// Health check kini memeriksa DUA database (PostgreSQL dan MongoDB)
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	// Auth routes
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// User management routes
	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	users.Patch("/:id/role", middleware.RequirePermission(deps.Permissions, "role:assign"), deps.AuthService.AssignRole)

	// Student routes
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	perms := deps.Permissions

	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequireJSON, middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)

	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", middleware.RequireJSON, deps.StudentService.Replace)
	students.Patch("/:id", middleware.RequireJSON, deps.StudentService.Patch)

	// Student Portfolio routes (Tugas Mandiri Modul 10)
	students.Get("/:id/portfolio", deps.PortfolioService.Get)
	students.Put("/:id/portfolio", middleware.RequireJSON, deps.PortfolioService.Upsert)
	students.Post("/:id/portfolio/projects", middleware.RequireJSON, deps.PortfolioService.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)

	// Global Portfolio stats
	portfolios := api.Group("/portfolios", middleware.RequireAuth(deps.JWT))
	portfolios.Get("/stats/technologies", deps.PortfolioService.StatsTechnologies)

	// Course routes
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Post("/swap", middleware.RequireJSON, deps.CourseService.Swap)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id/quota", middleware.RequireJSON, middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

	// Course Reviews routes (MongoDB)
	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
	courses.Post("/:id/reviews", middleware.RequireJSON, deps.ReviewService.Create)

	// Review detail & interaction routes (MongoDB)
	reviews := api.Group("/reviews", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete)
	reviews.Post("/:id/replies", deps.ReviewService.Reply)
	reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)

	// Supervisor routes
	supervisors := api.Group("/supervisors", middleware.RequireAuth(deps.JWT))
	supervisors.Get("/", deps.SupervisorService.List)
	supervisors.Post("/switch", middleware.RequireJSON, deps.SupervisorService.Switch)
	supervisors.Post("/:id/assign", middleware.RequireJSON, deps.SupervisorService.Assign)
	supervisors.Patch("/:id/quota", middleware.RequireJSON, middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.UpdateQuota)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.NewError(fiber.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.NewError(fiber.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "MongoDB tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
