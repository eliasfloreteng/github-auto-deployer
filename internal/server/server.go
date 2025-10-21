package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/eliasfloreteng/github-auto-deployer/internal/config"
	"github.com/eliasfloreteng/github-auto-deployer/internal/email"
	"github.com/eliasfloreteng/github-auto-deployer/internal/executor"
	"github.com/eliasfloreteng/github-auto-deployer/internal/git"
	githubpkg "github.com/eliasfloreteng/github-auto-deployer/internal/github"
	"github.com/google/go-github/v57/github"
)

// Server represents the webhook server
type Server struct {
	config      *config.Config
	emailClient *email.SMTPClient
}

// NewServer creates a new webhook server
func NewServer(cfg *config.Config) (*Server, error) {
	emailClient := email.NewSMTPClient(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.SMTP.To,
	)

	return &Server{
		config:      cfg,
		emailClient: emailClient,
	}, nil
}

// Start starts the webhook server
func (s *Server) Start() error {
	http.HandleFunc("/webhook", s.handleWebhook)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("Starting webhook server on %s", addr)

	return http.ListenAndServe(addr, nil)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// handleWebhook handles incoming GitHub webhooks
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		http.Error(w, "Error reading payload", http.StatusBadRequest)
		return
	}

	// Validate webhook signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !githubpkg.ValidateWebhookSignature(payload, signature, s.config.GitHub.WebhookSecret) {
		log.Printf("Invalid webhook signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse the webhook event
	event := r.Header.Get("X-GitHub-Event")
	if event != "push" {
		// We only care about push events
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse push event
	var pushEvent github.PushEvent
	if err := json.Unmarshal(payload, &pushEvent); err != nil {
		log.Printf("Error parsing push event: %v", err)
		http.Error(w, "Error parsing event", http.StatusBadRequest)
		return
	}

	// Handle the push event
	go s.handlePushEvent(&pushEvent)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processing",
	})
}

// handlePushEvent processes a push event
func (s *Server) handlePushEvent(event *github.PushEvent) {
	// Extract repository info
	repoOwner := event.GetRepo().GetOwner().GetLogin()
	repoName := event.GetRepo().GetName()
	branch := extractBranchFromRef(event.GetRef())

	log.Printf("Received push event for %s/%s on branch %s", repoOwner, repoName, branch)

	// Find matching watcher
	watcher, err := s.config.GetWatcherByRepo(repoOwner, repoName)
	if err != nil {
		log.Printf("No watcher found for %s/%s: %v", repoOwner, repoName, err)
		return
	}

	// Check if the branch matches
	if watcher.Branch != branch {
		log.Printf("Branch mismatch: watcher expects %s, got %s", watcher.Branch, branch)
		return
	}

	log.Printf("Processing deployment for %s (branch: %s)", watcher.Path, branch)

	// Create git manager
	gitMgr, err := git.NewManager(watcher.Path)
	if err != nil {
		log.Printf("Error creating git manager: %v", err)
		s.emailClient.SendFailureNotification(watcher.Path, branch, err.Error())
		return
	}

	// Fetch and pull changes
	if err := gitMgr.FetchAndPull(branch); err != nil {
		if git.IsConflictError(err) {
			log.Printf("Merge conflict detected: %v", err)
			s.emailClient.SendConflictNotification(watcher.Path, branch, err.Error())
		} else {
			log.Printf("Error pulling changes: %v", err)
			s.emailClient.SendFailureNotification(watcher.Path, branch, err.Error())
		}
		return
	}

	log.Printf("Successfully pulled changes for %s", watcher.Path)

	// Execute post-update command if specified
	if watcher.Command != "" {
		log.Printf("Executing command: %s", watcher.Command)
		exec := executor.NewExecutor(watcher.Path)
		if err := exec.Execute(watcher.Command); err != nil {
			log.Printf("Error executing command: %v", err)
			s.emailClient.SendCommandFailureNotification(watcher.Path, branch, watcher.Command, err.Error())
			return
		}
		log.Printf("Successfully executed command for %s", watcher.Path)
	}

	log.Printf("Deployment completed successfully for %s", watcher.Path)
}

// extractBranchFromRef extracts the branch name from a git ref
// e.g., "refs/heads/main" -> "main"
func extractBranchFromRef(ref string) string {
	const prefix = "refs/heads/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ref
}
