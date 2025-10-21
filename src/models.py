"""Data models for the GitHub Auto-Deployer."""
from typing import Optional
from pydantic import BaseModel, Field


class DeployerConfig(BaseModel):
    """Configuration from .deployer.yml file."""
    command: str = Field(..., description="Command to execute after successful pull")
    branch: Optional[str] = Field(None, description="Branch to watch (auto-detected if not specified)")
    notification_email: Optional[str] = Field(None, description="Email for notifications")
    command_timeout: int = Field(300, description="Command timeout in seconds")
    send_success_email: bool = Field(False, description="Send email on successful deployment")


class WatchedRepository(BaseModel):
    """Represents a watched repository."""
    path: str = Field(..., description="Absolute path to repository")
    remote_url: str = Field(..., description="Git remote URL")
    branch: str = Field(..., description="Branch being watched")
    config: DeployerConfig = Field(..., description="Deployer configuration")
    
    class Config:
        frozen = False


class GitHubWebhookPayload(BaseModel):
    """GitHub webhook push event payload."""
    ref: str = Field(..., description="Git ref (e.g., refs/heads/main)")
    repository: dict = Field(..., description="Repository information")
    pusher: dict = Field(..., description="Pusher information")
    commits: list = Field(default_factory=list, description="List of commits")
    
    @property
    def branch_name(self) -> str:
        """Extract branch name from ref."""
        if self.ref.startswith("refs/heads/"):
            return self.ref[11:]
        return self.ref
    
    @property
    def repo_url(self) -> str:
        """Get repository clone URL."""
        return self.repository.get("clone_url", "")
    
    @property
    def repo_ssh_url(self) -> str:
        """Get repository SSH URL."""
        return self.repository.get("ssh_url", "")
    
    @property
    def repo_html_url(self) -> str:
        """Get repository HTML URL."""
        return self.repository.get("html_url", "")


class DeploymentResult(BaseModel):
    """Result of a deployment operation."""
    success: bool
    repository_path: str
    branch: str
    message: str
    git_output: Optional[str] = None
    command_output: Optional[str] = None
    command_exit_code: Optional[int] = None
