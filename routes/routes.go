package routes

import (
	"christ-api/internal/auth"
	"christ-api/internal/contacts"
	"christ-api/internal/middleware"
	"christ-api/internal/news"
	"christ-api/internal/points"
	"christ-api/internal/role"
	"christ-api/internal/sites"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// public auth routes
	api.Post("/login", auth.Login)
	api.Post("/register", auth.Register)
	api.Post("/verify-otp", auth.VerifyOTP)
	api.Post("/auth/google", auth.LoginGoogle)
	api.Post("/auth/google/username", auth.SubmitGoogleUsername)

	// protected routes
	protected := api.Group("/", middleware.AuthMiddleware)

	protected.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "you are logged in",
		})
	})

	protected.Post("/logout", auth.Logout)

	// admin approvals
	protected.Get("/admin/approvals", auth.GetPendingApprovals)
	protected.Post("/admin/approvals/:id/approve", auth.ApproveUser)
	protected.Post("/admin/approvals/:id/reject", auth.RejectUser)

	// roles
	protected.Get("/roles", role.ListRoles)
	protected.Post("/roles", role.CreateRole)
	protected.Patch("/roles/:id", role.UpdateRole)

	// sites
	protected.Get("/sites", sites.ListSites)
	protected.Post("/sites", sites.CreateSite)
	protected.Patch("/sites/:uuid", sites.UpdateSite)

	// contacts
	protected.Get("/contacts", contacts.ListContacts)
	protected.Get("/contacts/:id", contacts.ListContacts)
	protected.Post("/contacts", contacts.CreateContact)
	protected.Patch("/contacts/:id", contacts.UpdateContact)
	protected.Delete("/contacts/:id", contacts.DeleteContact)

	// points
	protected.Get("/points", points.GetPoints)
	protected.Post("/points/earn", points.EarnPoints)
	protected.Post("/points/spend", points.SpendPoints)

	// news
	protected.Get("/news", news.ListNews)
	protected.Post("/news", news.CreateNews)
	protected.Patch("/news/:uuid", news.UpdateNews)
	protected.Delete("/news/:uuid", news.DeleteNews)
}
