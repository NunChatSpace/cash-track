package main

import (
	"log"
	"net/http"
	"net"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"cash-track/internal/config"
	"cash-track/internal/database"
	"cash-track/internal/handlers"
	"cash-track/internal/llm"
	"cash-track/internal/ocr"
	"cash-track/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	store, err := storage.NewLocalStorage(cfg.UploadDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	ocrClient := ocr.NewClient(cfg.OCREndpoint)
	llmClient := llm.NewClient(cfg.OllamaURL, cfg.OllamaModel)

	h, err := handlers.New(repo, store, ocrClient, llmClient, "web/templates")
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Pages
	r.Get("/", h.Index)
	r.Get("/chat", h.ChatPage)
	r.Get("/dashboard", h.DashboardPage)
	r.Get("/history", h.History)
	r.Get("/users", h.UsersPage)
	r.Get("/transactions/{id}/confirm", h.ConfirmPage)

	// API - Chat
	r.Post("/api/chat", h.Chat)

	// API - Transactions
	r.Post("/api/transactions/slip", h.UploadSlip)
	r.Get("/api/transactions/recent", h.GetRecentTransactions)
	r.Get("/api/transactions/{id}", h.GetTransaction)
	r.Patch("/api/transactions/{id}/confirm", h.ConfirmTransaction)
	r.Delete("/api/transactions/{id}", h.DeleteTransaction)

	// API - Users
	r.Get("/api/users", h.ListUsers)
	r.Post("/api/users", h.CreateUser)
	r.Post("/api/users/select", h.SelectUser)
	r.Patch("/api/users/{id}", h.UpdateCutoff)
	r.Delete("/api/users/{id}", h.DeleteUser)

	// API - Dashboard
	r.Get("/api/dashboard/summary", h.DashboardSummary)
	r.Get("/api/dashboard/by-category", h.DashboardByCategory)
	r.Get("/api/dashboard/by-channel", h.DashboardByChannel)

	log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
	for _, ip := range lanIPs() {
		log.Printf("LAN access: http://%s:%s", ip, cfg.ServerPort)
	}
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func lanIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}
