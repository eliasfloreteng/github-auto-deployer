"""Email notification service."""
import logging
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from typing import Optional
from datetime import datetime

logger = logging.getLogger(__name__)


class EmailNotifier:
    """Sends email notifications via SMTP."""
    
    def __init__(
        self,
        smtp_host: str,
        smtp_port: int,
        smtp_user: str,
        smtp_password: str,
        from_email: str,
    ):
        """
        Initialize EmailNotifier.
        
        Args:
            smtp_host: SMTP server hostname
            smtp_port: SMTP server port
            smtp_user: SMTP username
            smtp_password: SMTP password
            from_email: From email address
        """
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port
        self.smtp_user = smtp_user
        self.smtp_password = smtp_password
        self.from_email = from_email
    
    def send_email(
        self,
        to_email: str,
        subject: str,
        body: str,
        html: bool = False
    ) -> bool:
        """
        Send an email.
        
        Args:
            to_email: Recipient email address
            subject: Email subject
            body: Email body
            html: Whether body is HTML
            
        Returns:
            True if sent successfully, False otherwise
        """
        try:
            msg = MIMEMultipart('alternative')
            msg['From'] = self.from_email
            msg['To'] = to_email
            msg['Subject'] = subject
            msg['Date'] = datetime.now().strftime('%a, %d %b %Y %H:%M:%S %z')
            
            if html:
                msg.attach(MIMEText(body, 'html'))
            else:
                msg.attach(MIMEText(body, 'plain'))
            
            # Connect to SMTP server
            with smtplib.SMTP(self.smtp_host, self.smtp_port) as server:
                server.starttls()
                server.login(self.smtp_user, self.smtp_password)
                server.send_message(msg)
            
            logger.info(f"Email sent successfully to {to_email}")
            return True
            
        except Exception as e:
            logger.error(f"Failed to send email to {to_email}: {e}")
            return False
    
    def send_conflict_notification(
        self,
        to_email: str,
        repo_path: str,
        branch: str,
        conflict_details: str
    ) -> bool:
        """
        Send notification about git conflicts.
        
        Args:
            to_email: Recipient email
            repo_path: Repository path
            branch: Branch name
            conflict_details: Details about the conflict
            
        Returns:
            True if sent successfully
        """
        subject = f"⚠️ Git Conflict Detected - {repo_path}"
        
        body = f"""
GitHub Auto-Deployer - Conflict Notification

Repository: {repo_path}
Branch: {branch}
Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

A git pull operation failed due to conflicts or uncommitted changes.

Details:
{conflict_details}

Action Required:
Please manually resolve the conflicts in the repository and ensure all changes are committed.
Once resolved, the next push will trigger automatic deployment again.

---
This is an automated message from GitHub Auto-Deployer.
"""
        
        return self.send_email(to_email, subject, body)
    
    def send_pull_failure_notification(
        self,
        to_email: str,
        repo_path: str,
        branch: str,
        error_message: str
    ) -> bool:
        """
        Send notification about git pull failure.
        
        Args:
            to_email: Recipient email
            repo_path: Repository path
            branch: Branch name
            error_message: Error message
            
        Returns:
            True if sent successfully
        """
        subject = f"❌ Git Pull Failed - {repo_path}"
        
        body = f"""
GitHub Auto-Deployer - Pull Failure Notification

Repository: {repo_path}
Branch: {branch}
Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

Failed to pull the latest changes from the remote repository.

Error:
{error_message}

Action Required:
Please check the repository and resolve any issues preventing the pull operation.

---
This is an automated message from GitHub Auto-Deployer.
"""
        
        return self.send_email(to_email, subject, body)
    
    def send_command_failure_notification(
        self,
        to_email: str,
        repo_path: str,
        branch: str,
        command: str,
        exit_code: int,
        output: str
    ) -> bool:
        """
        Send notification about command execution failure.
        
        Args:
            to_email: Recipient email
            repo_path: Repository path
            branch: Branch name
            command: Command that failed
            exit_code: Command exit code
            output: Command output
            
        Returns:
            True if sent successfully
        """
        subject = f"❌ Deployment Command Failed - {repo_path}"
        
        body = f"""
GitHub Auto-Deployer - Command Failure Notification

Repository: {repo_path}
Branch: {branch}
Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

The deployment command failed after successfully pulling the latest changes.

Command: {command}
Exit Code: {exit_code}

Output:
{output}

Action Required:
Please check the command configuration and repository state.

---
This is an automated message from GitHub Auto-Deployer.
"""
        
        return self.send_email(to_email, subject, body)
    
    def send_success_notification(
        self,
        to_email: str,
        repo_path: str,
        branch: str,
        command: str,
        commit_info: Optional[str] = None
    ) -> bool:
        """
        Send notification about successful deployment.
        
        Args:
            to_email: Recipient email
            repo_path: Repository path
            branch: Branch name
            command: Command that was executed
            commit_info: Information about the commit
            
        Returns:
            True if sent successfully
        """
        subject = f"✅ Deployment Successful - {repo_path}"
        
        commit_section = ""
        if commit_info:
            commit_section = f"\nCommit:\n{commit_info}\n"
        
        body = f"""
GitHub Auto-Deployer - Success Notification

Repository: {repo_path}
Branch: {branch}
Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

Successfully pulled the latest changes and executed the deployment command.
{commit_section}
Command: {command}

---
This is an automated message from GitHub Auto-Deployer.
"""
        
        return self.send_email(to_email, subject, body)
