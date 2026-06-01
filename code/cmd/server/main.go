package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"blog-engine/config"
	"blog-engine/internal/auth"
	"blog-engine/internal/blog"
	"blog-engine/internal/middleware"
	"blog-engine/internal/notification"
	"blog-engine/internal/search"
	"blog-engine/internal/social"
	"blog-engine/internal/upload"
	"blog-engine/internal/user"
	"blog-engine/internal/admin"
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
	blogSvc    := blog.NewService(blogRepo, sanitizer)
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

	// Handlers
	authH   := auth.NewHandler(authSvc)
	blogH   := blog.NewHandler(blogSvc)
	uploadH := upload.NewHandler(uploadSvc)
	socialH := social.NewHandler(socialSvc)
	notifH  := notification.NewHandler(notifSvc)
	searchH := search.NewHandler(searchSvc)
	userH   := user.NewHandler(userSvc)
	adminH  := admin.NewHandler(adminSvc)

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
		r.Post("/auth/logout", authH.Logout)

		// Blogs (public read)
		r.Get("/blogs/feed/explore", blogH.ExploreFeed)
		r.Get("/blogs/{id}", blogH.GetBlog)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)

			r.Get("/blogs/feed/following", blogH.FollowingFeed)
			r.Post("/blogs", blogH.CreateBlog)
			r.Patch("/blogs/{id}", blogH.UpdateBlog)
			r.Delete("/blogs/{id}", blogH.DeleteBlog)

			r.Post("/uploads/image", uploadH.UploadImage)

			r.Get("/search", searchH.Search)

			r.Get("/users/{username}", userH.GetProfile)
			r.Patch("/users/me", userH.UpdateProfile)

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
