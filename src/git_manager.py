"""Git operations manager."""
import logging
from pathlib import Path
from typing import Tuple, Optional
import git
from git.exc import GitCommandError

logger = logging.getLogger(__name__)


class GitManager:
    """Manages git operations for repositories."""
    
    def __init__(self, repo_path: Path):
        """
        Initialize GitManager for a repository.
        
        Args:
            repo_path: Path to the git repository
        """
        self.repo_path = repo_path
        try:
            self.repo = git.Repo(repo_path)
        except git.InvalidGitRepositoryError:
            logger.error(f"Invalid git repository: {repo_path}")
            raise
    
    def get_current_branch(self) -> str:
        """
        Get the currently checked out branch name.
        
        Returns:
            Branch name
        """
        try:
            return self.repo.active_branch.name
        except Exception as e:
            logger.error(f"Error getting current branch for {self.repo_path}: {e}")
            return "unknown"
    
    def get_remote_url(self) -> str:
        """
        Get the remote URL for origin.
        
        Returns:
            Remote URL or empty string if not found
        """
        try:
            if 'origin' in self.repo.remotes:
                return self.repo.remotes.origin.url
            return ""
        except Exception as e:
            logger.error(f"Error getting remote URL for {self.repo_path}: {e}")
            return ""
    
    def normalize_remote_url(self, url: str) -> str:
        """
        Normalize a git remote URL for comparison.
        Handles both HTTPS and SSH URLs.
        
        Args:
            url: Git remote URL
            
        Returns:
            Normalized URL
        """
        # Remove .git suffix
        url = url.rstrip('/')
        if url.endswith('.git'):
            url = url[:-4]
        
        # Convert SSH to HTTPS format for comparison
        if url.startswith('git@github.com:'):
            url = url.replace('git@github.com:', 'https://github.com/')
        
        return url.lower()
    
    def matches_remote(self, remote_url: str) -> bool:
        """
        Check if the given remote URL matches this repository's remote.
        
        Args:
            remote_url: Remote URL to compare
            
        Returns:
            True if URLs match, False otherwise
        """
        local_url = self.normalize_remote_url(self.get_remote_url())
        compare_url = self.normalize_remote_url(remote_url)
        return local_url == compare_url
    
    def has_conflicts(self) -> Tuple[bool, str]:
        """
        Check if the repository has any conflicts or uncommitted changes.
        
        Returns:
            Tuple of (has_conflicts, status_message)
        """
        try:
            # Check for uncommitted changes
            if self.repo.is_dirty(untracked_files=False):
                status = self.repo.git.status()
                return True, f"Repository has uncommitted changes:\n{status}"
            
            # Check for untracked files that might conflict
            untracked = self.repo.untracked_files
            if untracked:
                logger.warning(f"Repository has untracked files: {untracked}")
            
            return False, "Repository is clean"
            
        except Exception as e:
            logger.error(f"Error checking repository status: {e}")
            return True, f"Error checking status: {str(e)}"
    
    def fetch_and_pull(self, branch: str) -> Tuple[bool, str]:
        """
        Fetch and pull the latest changes from the remote branch.
        
        Args:
            branch: Branch name to pull
            
        Returns:
            Tuple of (success, output_message)
        """
        output_lines = []
        
        try:
            # Check for conflicts first
            has_conflicts, conflict_msg = self.has_conflicts()
            if has_conflicts:
                logger.error(f"Cannot pull - {conflict_msg}")
                return False, conflict_msg
            
            # Fetch from remote
            logger.info(f"Fetching from origin for {self.repo_path}")
            fetch_info = self.repo.remotes.origin.fetch()
            output_lines.append(f"Fetched from origin: {[str(info) for info in fetch_info]}")
            
            # Check if we're on the correct branch
            current_branch = self.get_current_branch()
            if current_branch != branch:
                logger.warning(
                    f"Current branch '{current_branch}' doesn't match target branch '{branch}'"
                )
                output_lines.append(
                    f"Warning: On branch '{current_branch}', expected '{branch}'"
                )
            
            # Pull changes
            logger.info(f"Pulling branch '{branch}' for {self.repo_path}")
            pull_info = self.repo.remotes.origin.pull(branch)
            output_lines.append(f"Pull result: {[str(info) for info in pull_info]}")
            
            # Get the latest commit info
            latest_commit = self.repo.head.commit
            output_lines.append(
                f"Latest commit: {latest_commit.hexsha[:7]} - {latest_commit.message.strip()}"
            )
            
            success_msg = "\n".join(output_lines)
            logger.info(f"Successfully pulled changes for {self.repo_path}")
            return True, success_msg
            
        except GitCommandError as e:
            error_msg = f"Git command failed: {e.stderr if e.stderr else str(e)}"
            logger.error(f"Pull failed for {self.repo_path}: {error_msg}")
            return False, error_msg
            
        except Exception as e:
            error_msg = f"Unexpected error during pull: {str(e)}"
            logger.error(f"Pull failed for {self.repo_path}: {error_msg}")
            return False, error_msg
    
    def get_repository_info(self) -> dict:
        """
        Get information about the repository.
        
        Returns:
            Dictionary with repository information
        """
        try:
            return {
                "path": str(self.repo_path),
                "branch": self.get_current_branch(),
                "remote_url": self.get_remote_url(),
                "latest_commit": self.repo.head.commit.hexsha[:7],
                "is_dirty": self.repo.is_dirty(),
            }
        except Exception as e:
            logger.error(f"Error getting repository info: {e}")
            return {
                "path": str(self.repo_path),
                "error": str(e)
            }
