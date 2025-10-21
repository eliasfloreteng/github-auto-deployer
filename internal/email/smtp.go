package email

import (
	"fmt"
	"time"

	"gopkg.in/gomail.v2"
)

// SMTPClient handles email notifications
type SMTPClient struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       string
}

// NewSMTPClient creates a new SMTP client
func NewSMTPClient(host string, port int, username, password, from, to string) *SMTPClient {
	return &SMTPClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

// SendFailureNotification sends an email notification about a deployment failure
func (s *SMTPClient) SendFailureNotification(repoPath, branch, errorMsg string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", s.to)
	m.SetHeader("Subject", fmt.Sprintf("GitHub Auto-Deployer: Deployment Failed for %s", repoPath))

	body := fmt.Sprintf(`
Deployment Failure Notification
================================

Repository: %s
Branch: %s
Time: %s

Error Details:
--------------
%s

Please check the repository and resolve any conflicts or issues.

---
This is an automated message from GitHub Auto-Deployer.
`, repoPath, branch, time.Now().Format(time.RFC1123), errorMsg)

	m.SetBody("text/plain", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendConflictNotification sends an email notification about merge conflicts
func (s *SMTPClient) SendConflictNotification(repoPath, branch, conflictDetails string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", s.to)
	m.SetHeader("Subject", fmt.Sprintf("GitHub Auto-Deployer: Merge Conflict in %s", repoPath))

	body := fmt.Sprintf(`
Merge Conflict Detected
=======================

Repository: %s
Branch: %s
Time: %s

Conflict Details:
-----------------
%s

Action Required:
----------------
Please manually resolve the conflicts in the repository and commit the changes.
The auto-deployer will not be able to update this repository until conflicts are resolved.

---
This is an automated message from GitHub Auto-Deployer.
`, repoPath, branch, time.Now().Format(time.RFC1123), conflictDetails)

	m.SetBody("text/plain", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendCommandFailureNotification sends an email notification about command execution failure
func (s *SMTPClient) SendCommandFailureNotification(repoPath, branch, command, errorMsg string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", s.to)
	m.SetHeader("Subject", fmt.Sprintf("GitHub Auto-Deployer: Command Failed for %s", repoPath))

	body := fmt.Sprintf(`
Command Execution Failure
=========================

Repository: %s
Branch: %s
Command: %s
Time: %s

Error Details:
--------------
%s

The repository was updated successfully, but the post-update command failed.
Please check the command and repository state.

---
This is an automated message from GitHub Auto-Deployer.
`, repoPath, branch, command, time.Now().Format(time.RFC1123), errorMsg)

	m.SetBody("text/plain", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
