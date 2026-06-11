package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"blog-engine/config"
	"blog-engine/internal/admin"
	"blog-engine/internal/auth"
	"blog-engine/internal/blog"
	"blog-engine/internal/middleware"
	"blog-engine/internal/notification"
	"blog-engine/internal/portfolio"
	"blog-engine/internal/search"
	"blog-engine/internal/site"
	"blog-engine/internal/social"
	"blog-engine/internal/translation"
	"blog-engine/internal/upload"
	"blog-engine/internal/user"
	"blog-engine/pkg/database"
	"blog-engine/pkg/email"
	"blog-engine/pkg/sanitize"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	// Infrastructure
	sanitizer  := sanitize.NewHTMLSanitizer()
	emailSvc   := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, cfg.AppURL)
	r2Client   := upload.NewR2Client(cfg.R2AccountID, cfg.R2AccessKey, cfg.R2SecretKey, cfg.R2BucketName, cfg.R2PublicURL)
	jwtUtil    := auth.NewJWT(cfg.JWTSecret, cfg.JWTRefreshSecret)

	// Repositories
	authRepo   := auth.NewPostgresRepository(db)

	// Services
	authSvc    := auth.NewService(authRepo, emailSvc, cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.AppURL)
	blogRepo   := blog.NewPostgresRepository(db)
	var translator blog.Translator
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		translator = translation.NewClaudeTranslator(apiKey)
	}
	blogSvc    := blog.NewService(blogRepo, sanitizer, translator)
	uploadSvc  := upload.NewService(r2Client, cfg.R2PublicURL)
	notifRepo  := notification.NewPostgresRepository(db)
	notifSvc   := notification.NewService(notifRepo)
	socialRepo := social.NewPostgresRepository(db)
	socialSvc  := social.NewService(socialRepo, &notifBridge{svc: notifSvc})
	searchRepo := search.NewPostgresRepository(db)
	searchSvc  := search.NewService(searchRepo)
	userRepo   := user.NewPostgresRepository(db)
	userSvc    := user.NewService(userRepo)
	adminRepo  := admin.NewPostgresRepository(db)
	adminSvc   := admin.NewService(adminRepo)
	portRepo   := portfolio.NewPostgresRepository(db)
	portSvc    := portfolio.NewService(portRepo, sanitizer)
	siteRepo   := site.NewPostgresRepository(db)
	siteSvc    := site.NewService(siteRepo, sanitizer)

	// Handlers
	authH   := auth.NewHandler(authSvc)
	blogH   := blog.NewHandler(blogSvc)
	uploadH := upload.NewHandler(uploadSvc)
	socialH := social.NewHandler(socialSvc)
	notifH  := notification.NewHandler(notifSvc)
	searchH := search.NewHandler(searchSvc)
	userH   := user.NewHandler(userSvc)
	adminH  := admin.NewHandler(adminSvc)
	portH   := portfolio.NewHandler(portSvc)
	siteH   := site.NewHandler(siteSvc)

	// Middleware
	authMw := middleware.NewAuth(jwtUtil)
	rbacMw := middleware.NewRBAC()

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(corsMiddleware(cfg.AppURL))

	r.Get("/health", auth.HealthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth (public)
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Get("/auth/verify", authH.VerifyEmail)
		r.Post("/auth/forgot-password", authH.ForgotPassword)
		r.Post("/auth/reset-password", authH.ResetPassword)
		r.Post("/auth/refresh", authH.Refresh)
		r.Post("/auth/logout", authH.Logout)

		// Blogs (public read — guests allowed; a presented token is validated
		// at the routing level and rejected with 401 if invalid)
		r.Group(func(r chi.Router) {
			r.Use(authMw.OptionalAuthenticate)
			r.Get("/blogs/feed", blogH.ArticlesFeed)
			r.Get("/blogs/{id}", blogH.GetBlog)
		})

		// CR-002 public pages
		r.Get("/categories", blogH.ListCategories)
		r.Get("/projects", portH.ListProjects)
		r.Get("/about", siteH.GetAbout)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)

			r.Patch("/blogs/{id}", blogH.UpdateBlog)
			r.Delete("/blogs/{id}", blogH.DeleteBlog)

			r.Post("/uploads/image", uploadH.UploadImage)

			r.Get("/search", searchH.Search)

			r.Get("/users/{username}", userH.GetProfile)
			r.Patch("/users/me", userH.UpdateProfile)
				r.Put("/users/me/language", userH.UpdateLanguagePreference)

			r.Post("/users/{id}/follow", socialH.Follow)
			r.Delete("/users/{id}/follow", socialH.Unfollow)
			r.Post("/users/{id}/friend-request", socialH.SendFriendRequest)
			r.Patch("/friend-requests/{id}", socialH.RespondFriendRequest)
			r.Delete("/users/{id}/friend", socialH.Unfriend)
			r.Post("/users/{id}/block", socialH.BlockUser)
			r.Delete("/users/{id}/block", socialH.UnblockUser)

			r.Post("/blogs/{id}/react", socialH.React)
			r.Delete("/blogs/{id}/react", socialH.RemoveReaction)
			r.Post("/blogs/{id}/comments", socialH.CreateComment)
			r.Delete("/comments/{id}", socialH.DeleteComment)

			r.Post("/reports", socialH.Report)

			r.Get("/notifications", notifH.List)
			r.Patch("/notifications/{id}/read", notifH.MarkRead)
			r.Patch("/notifications/read-all", notifH.MarkAllRead)
		})

		// Owner-only routes (CR-001 personal-blog pivot: only the owner writes)
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(rbacMw.RequireRole("owner"))
			r.Post("/blogs", blogH.CreateBlog)
			// CR-002: portfolio + author page management
			r.Post("/projects", portH.CreateProject)
			r.Patch("/projects/{id}", portH.UpdateProject)
			r.Delete("/projects/{id}", portH.DeleteProject)
			r.Put("/about", siteH.UpdateAbout)
		})

		// Moderator+ routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(rbacMw.RequireRole("moderator", "admin", "owner"))
			r.Get("/admin/reports", adminH.ListReports)
			r.Patch("/admin/reports/{id}/resolve", adminH.ResolveReport)
		})

		// Admin+ routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Use(rbacMw.RequireRole("admin", "owner"))
			r.Get("/admin/users", adminH.ListUsers)
			r.Patch("/admin/users/{id}/role", adminH.ChangeRole)
			r.Get("/admin/stats", adminH.GetStats)
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
	}
}

// notifBridge adapts notification.Service to social.Notifier — avoids circular import
type notifBridge struct{ svc *notification.Service }

func (n *notifBridge) Notify(ctx context.Context, input *social.NotifyInput) error {
	ni := &notification.CreateInput{
		Type:        input.Type,
		ActorID:     input.ActorID,
		RecipientID: input.RecipientID,
		BlogID:      input.BlogID,
		CommentID:   input.CommentID,
	}
	if input.BroadcastToMods {
		return n.svc.BroadcastToMods(ctx, ni)
	}
	return n.svc.Create(ctx, ni)
}

func corsMiddleware(appURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", appURL)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
