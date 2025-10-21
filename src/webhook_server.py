"""FastAPI webhook server for GitHub events."""
import logging
import hmac
import hashlib
from typing import Optional
from fastapi import FastAPI, Request, HTTPException, BackgroundTasks
from fastapi.responses import JSONResponse

from .models import GitHubWebhookPayload, DeploymentResult
from .folder_monitor import FolderMonitor
from .git_manager import GitManager
from .command_executor import CommandExecutor
from .email_notifier import EmailNotifier
from pathlib import Path

logger = logging.getLogger(__name__)


class WebhookServer:
    """FastAPI server for handling GitHub webhooks."""
    
    def __init__(
        self,
        folder_monitor: FolderMonitor,
        email_notifier: EmailNotifier,
        webhook_secret: str
    ):
        """
        Initialize WebhookServer.
        
        Args:
            folder_monitor: FolderMonitor instance
            email_notifier: EmailNotifier instance
            webhook_secret: GitHub webhook secret for signature validation
        """
        self.folder_monitor = folder_monitor
        self.email_notifier = email_notifier
        self.webhook_secret = webhook_secret
        self.app = FastAPI(title="GitHub Auto-Deployer")
        
        # Register routes
        self._setup_routes()
    
    def _setup_routes(self):
        """Setup FastAPI routes."""
        
        @self.app.get("/")
        async def root():
            """Root endpoint."""
            return {
                "service": "GitHub Auto-Deployer",
                "status": "running",
                "watched_repositories": self.folder_monitor.get_repository_count()
            }
        
        @self.app.get("/health")
        async def health():
            """Health check endpoint."""
            return {
                "status": "healthy",
                "watched_repositories": self.folder_monitor.get_repository_count()
            }
        
        @self.app.get("/repositories")
        async def list_repositories():
            """List all watched repositories."""
            repos = self.folder_monitor.get_all_repositories()
            return {
                "count": len(repos),
                "repositories": [
                    {
                        "path": repo.path,
                        "remote_url": repo.remote_url,
                        "branch": repo.branch,
                        "command": repo.config.command
                    }
                    for repo in repos.values()
                ]
            }
        
        @self.app.post("/webhook")
        async def webhook(request: Request, background_tasks: BackgroundTasks):
            """
            Handle GitHub webhook events.
            
            Args:
                request: FastAPI request object
                background_tasks: Background tasks for async processing
                
            Returns:
                JSON response
            """
            # Get headers
            signature = request.headers.get("X-Hub-Signature-256")
            event_type = request.headers.get("X-GitHub-Event")
            
            # Read body
            body = await request.body()
            
            # Validate signature
            if not self._verify_signature(body, signature):
                logger.warning("Invalid webhook signature")
                raise HTTPException(status_code=401, detail="Invalid signature")
            
            # Only process push events
            if event_type != "push":
                logger.info(f"Ignoring event type: {event_type}")
                return {"status": "ignored", "reason": f"Event type '{event_type}' not supported"}
            
            # Parse payload
            try:
                import json
                payload_dict = json.loads(body)
                payload = GitHubWebhookPayload(**payload_dict)
            except Exception as e:
                logger.error(f"Error parsing webhook payload: {e}")
                raise HTTPException(status_code=400, detail="Invalid payload")
            
            # Process webhook in background
            background_tasks.add_task(
                self._process_push_event,
                payload
            )
            
            return {
                "status": "accepted",
                "repository": payload.repository.get("full_name"),
                "branch": payload.branch_name
            }
    
    def _verify_signature(self, payload: bytes, signature: Optional[str]) -> bool:
        """
        Verify GitHub webhook signature.
        
        Args:
            payload: Request body
            signature: X-Hub-Signature-256 header value
            
        Returns:
            True if signature is valid
        """
        if not signature:
            logger.warning("No signature provided")
            return False
        
        if not signature.startswith("sha256="):
            logger.warning("Invalid signature format")
            return False
        
        # Calculate expected signature
        expected_signature = "sha256=" + hmac.new(
            self.webhook_secret.encode(),
            payload,
            hashlib.sha256
        ).hexdigest()
        
        # Compare signatures
        return hmac.compare_digest(expected_signature, signature)
    
    async def _process_push_event(self, payload: GitHubWebhookPayload):
        """
        Process a push event.
        
        Args:
            payload: GitHub webhook payload
        """
        logger.info(
            f"Processing push event for {payload.repository.get('full_name')} "
            f"on branch {payload.branch_name}"
        )
        
        # Find matching repository
        repo = self.folder_monitor.find_repository_by_remote(
            payload.repo_url,
            payload.branch_name
        )
        
        if not repo:
            # Try SSH URL
            repo = self.folder_monitor.find_repository_by_remote(
                payload.repo_ssh_url,
                payload.branch_name
            )
        
        if not repo:
            logger.info(
                f"No matching repository found for {payload.repo_url} "
                f"on branch {payload.branch_name}"
            )
            return
        
        logger.info(f"Found matching repository: {repo.path}")
        
        # Execute deployment
        result = await self._deploy_repository(repo, payload)
        
        # Log result
        if result.success:
            logger.info(f"Deployment successful for {repo.path}")
        else:
            logger.error(f"Deployment failed for {repo.path}: {result.message}")
    
    async def _deploy_repository(
        self,
        repo,
        payload: GitHubWebhookPayload
    ) -> DeploymentResult:
        """
        Deploy a repository.
        
        Args:
            repo: WatchedRepository instance
            payload: GitHub webhook payload
            
        Returns:
            DeploymentResult
        """
        repo_path = Path(repo.path)
        
        try:
            # Initialize git manager
            git_manager = GitManager(repo_path)
            
            # Fetch and pull
            success, git_output = git_manager.fetch_and_pull(repo.branch)
            
            if not success:
                # Send failure notification
                self.email_notifier.send_pull_failure_notification(
                    to_email=repo.config.notification_email,
                    repo_path=repo.path,
                    branch=repo.branch,
                    error_message=git_output
                )
                
                return DeploymentResult(
                    success=False,
                    repository_path=repo.path,
                    branch=repo.branch,
                    message="Git pull failed",
                    git_output=git_output
                )
            
            logger.info(f"Successfully pulled changes for {repo.path}")
            
            # Execute deployment command
            executor = CommandExecutor(repo_path)
            cmd_success, cmd_output, exit_code = executor.execute_safe(
                repo.config.command,
                repo.config.command_timeout
            )
            
            if not cmd_success:
                # Send command failure notification
                self.email_notifier.send_command_failure_notification(
                    to_email=repo.config.notification_email,
                    repo_path=repo.path,
                    branch=repo.branch,
                    command=repo.config.command,
                    exit_code=exit_code,
                    output=cmd_output
                )
                
                return DeploymentResult(
                    success=False,
                    repository_path=repo.path,
                    branch=repo.branch,
                    message="Command execution failed",
                    git_output=git_output,
                    command_output=cmd_output,
                    command_exit_code=exit_code
                )
            
            logger.info(f"Successfully executed command for {repo.path}")
            
            # Send success notification if configured
            if repo.config.send_success_email:
                commit_info = None
                if payload.commits:
                    latest_commit = payload.commits[-1]
                    commit_info = f"{latest_commit.get('id', '')[:7]} - {latest_commit.get('message', '')}"
                
                self.email_notifier.send_success_notification(
                    to_email=repo.config.notification_email,
                    repo_path=repo.path,
                    branch=repo.branch,
                    command=repo.config.command,
                    commit_info=commit_info
                )
            
            return DeploymentResult(
                success=True,
                repository_path=repo.path,
                branch=repo.branch,
                message="Deployment successful",
                git_output=git_output,
                command_output=cmd_output,
                command_exit_code=exit_code
            )
            
        except Exception as e:
            error_msg = f"Unexpected error during deployment: {str(e)}"
            logger.error(error_msg)
            
            # Send error notification
            self.email_notifier.send_pull_failure_notification(
                to_email=repo.config.notification_email,
                repo_path=repo.path,
                branch=repo.branch,
                error_message=error_msg
            )
            
            return DeploymentResult(
                success=False,
                repository_path=repo.path,
                branch=repo.branch,
                message=error_msg
            )
    
    def get_app(self) -> FastAPI:
        """
        Get the FastAPI application.
        
        Returns:
            FastAPI app instance
        """
        return self.app
