package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/eliasfloreteng/github-auto-deployer/internal/config"
	"github.com/eliasfloreteng/github-auto-deployer/internal/executor"
	"github.com/eliasfloreteng/github-auto-deployer/internal/git"
	"github.com/eliasfloreteng/github-auto-deployer/internal/notifier"
)

// Handler handles GitHub webhook requests
type Handler struct {
	config   *config.Config
	notifier *notifier.EmailNotifier
}

// NewHandler creates a new webhook handler
func NewHandler(cfg *config.Config) *Handler {
	emailNotifier := notifier.NewEmailNotifier(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.SMTP.To,
	)

	return &Handler{
		config:   cfg,
		notifier: emailNotifier,
	}
}

// ServeHTTP handles incoming webhook requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySignature(body, signature) {
		log.Printf("Invalid webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		// We only care about push events
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse push event
	var pushEvent PushEvent
	if err := json.Unmarshal(body, &pushEvent); err != nil {
		log.Printf("Error parsing push event: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process the push event
	go h.processPushEvent(&pushEvent)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// verifySignature verifies the GitHub webhook signature
func (h *Handler) verifySignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Remove "sha256=" prefix
	signature = strings.TrimPrefix(signature, "sha256=")

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(h.config.GitHub.WebhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// processPushEvent processes a push event
func (h *Handler) processPushEvent(event *PushEvent) {
	log.Printf("Processing push event for %s, branch: %s", event.Repository.FullName, event.Ref)

	// Extract branch name from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	// Find matching watched folders
	for _, folder := range h.config.Folders {
		// Check if repository URL matches
		if !git.CompareURLs(folder.RepoURL, event.Repository.CloneURL) {
			continue
		}

		// Check if branch matches
		if folder.Branch != branch {
			log.Printf("Branch mismatch for %s: expected %s, got %s", folder.Path, folder.Branch, branch)
			continue
		}

		log.Printf("Matched folder: %s", folder.Path)

		// Process the update
		if err := h.processUpdate(&folder); err != nil {
			log.Printf("Error processing update for %s: %v", folder.Path, err)
			// Send failure notification
			if err := h.notifier.SendFailureNotification(folder.Path, branch, err.Error()); err != nil {
				log.Printf("Error sending failure notification: %v", err)
			}
		} else {
			log.Printf("Successfully processed update for %s", folder.Path)
		}
	}
}

// processUpdate handles the git pull and command execution
func (h *Handler) processUpdate(folder *config.WatchedFolder) error {
	// Create git manager
	gitMgr := git.NewManager(folder.Path)

	// Pull latest changes
	log.Printf("Pulling latest changes for %s", folder.Path)
	if err := gitMgr.Pull(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	// Execute post-update command
	if folder.Command != "" {
		log.Printf("Executing command for %s: %s", folder.Path, folder.Command)
		exec := executor.NewExecutor(folder.Path)
		output, err := exec.Execute(folder.Command)
		if err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}
		log.Printf("Command output: %s", output)
	}

	return nil
}

// PushEvent represents a GitHub push event
type PushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}
