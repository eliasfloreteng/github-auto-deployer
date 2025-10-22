package notifier

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// EmailNotifier handles email notifications
type EmailNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       string
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(host string, port int, username, password, from, to string) *EmailNotifier {
	return &EmailNotifier{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

// SendFailureNotification sends an email notification about a deployment failure
func (n *EmailNotifier) SendFailureNotification(repoPath, branch, errorMsg string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", n.from)
	m.SetHeader("To", n.to)
	m.SetHeader("Subject", fmt.Sprintf("Deployment Failed: %s", repoPath))

	body := fmt.Sprintf(`
Deployment Failure Notification

Repository: %s
Branch: %s
Time: %s

Error:
%s

Please check the repository and resolve any conflicts manually.
`, repoPath, branch, getCurrentTime(), errorMsg)

	m.SetBody("text/plain", body)

	d := gomail.NewDialer(n.host, n.port, n.username, n.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// getCurrentTime returns the current time as a string
func getCurrentTime() string {
	return fmt.Sprintf("%v", gomail.NewMessage().GetHeader("Date"))
}
