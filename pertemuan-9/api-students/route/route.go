package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-students-v9/app/service"
	"api-students-v9/helper"
	"api-students-v9/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	Permissions       *helper.PermissionSet
	StudentService    *service.StudentService
	AuthService       *service.AuthService
	CourseService     *service.CourseService
	SupervisorService *service.SupervisorService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck(deps.Pool))

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

	// Course routes
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Post("/swap", middleware.RequireJSON, deps.CourseService.Swap)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id/quota", middleware.RequireJSON, middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

	// Supervisor routes
	supervisors := api.Group("/supervisors", middleware.RequireAuth(deps.JWT))
	supervisors.Get("/", deps.SupervisorService.List)
	supervisors.Post("/switch", middleware.RequireJSON, deps.SupervisorService.Switch)
	supervisors.Post("/:id/assign", middleware.RequireJSON, deps.SupervisorService.Assign)
	supervisors.Patch("/:id/quota", middleware.RequireJSON, middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.UpdateQuota)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.NewError(fiber.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
