"""Folder monitoring service using watchdog."""
import logging
from pathlib import Path
from typing import Dict, Optional, Callable
import time
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler, FileSystemEvent

from .config_parser import ConfigParser
from .git_manager import GitManager
from .models import WatchedRepository, DeployerConfig

logger = logging.getLogger(__name__)


class RepositoryEventHandler(FileSystemEventHandler):
    """Handles filesystem events for repository monitoring."""
    
    def __init__(self, on_config_added: Callable, on_config_removed: Callable):
        """
        Initialize event handler.
        
        Args:
            on_config_added: Callback when config file is added
            on_config_removed: Callback when config file is removed
        """
        super().__init__()
        self.on_config_added = on_config_added
        self.on_config_removed = on_config_removed
        self.config_filename = ConfigParser.CONFIG_FILENAME
    
    def on_created(self, event: FileSystemEvent):
        """Handle file creation events."""
        if event.is_directory:
            return
        
        path = Path(event.src_path)
        if path.name == self.config_filename:
            logger.info(f"Config file created: {path}")
            self.on_config_added(path.parent)
    
    def on_deleted(self, event: FileSystemEvent):
        """Handle file deletion events."""
        if event.is_directory:
            return
        
        path = Path(event.src_path)
        if path.name == self.config_filename:
            logger.info(f"Config file deleted: {path}")
            self.on_config_removed(path.parent)
    
    def on_modified(self, event: FileSystemEvent):
        """Handle file modification events."""
        if event.is_directory:
            return
        
        path = Path(event.src_path)
        if path.name == self.config_filename:
            logger.info(f"Config file modified: {path}")
            # Treat modification as remove + add to reload config
            self.on_config_removed(path.parent)
            self.on_config_added(path.parent)


class FolderMonitor:
    """Monitors folders for git repositories with deployer configs."""
    
    def __init__(self, root_path: Path, default_email: str):
        """
        Initialize FolderMonitor.
        
        Args:
            root_path: Root path to monitor
            default_email: Default notification email
        """
        self.root_path = root_path
        self.default_email = default_email
        self.watched_repos: Dict[str, WatchedRepository] = {}
        self.observer: Optional[Observer] = None
        
        logger.info(f"Initializing FolderMonitor for {root_path}")
    
    def _register_repository(self, repo_path: Path) -> bool:
        """
        Register a repository for monitoring.
        
        Args:
            repo_path: Path to the repository
            
        Returns:
            True if registered successfully
        """
        try:
            # Check if it's a valid repository with config
            if not ConfigParser.is_valid_repository(repo_path):
                return False
            
            # Parse config
            config_path = ConfigParser.find_config(repo_path)
            if not config_path:
                return False
            
            config = ConfigParser.parse_config(config_path)
            if not config:
                return False
            
            # Get git information
            git_manager = GitManager(repo_path)
            remote_url = git_manager.get_remote_url()
            
            if not remote_url:
                logger.warning(f"No remote URL found for {repo_path}")
                return False
            
            # Determine branch
            branch = config.branch if config.branch else git_manager.get_current_branch()
            
            # Determine notification email
            notification_email = config.notification_email if config.notification_email else self.default_email
            
            # Update config with resolved values
            config.notification_email = notification_email
            if not config.branch:
                config.branch = branch
            
            # Create watched repository
            watched_repo = WatchedRepository(
                path=str(repo_path),
                remote_url=remote_url,
                branch=branch,
                config=config
            )
            
            # Store in registry using path as key
            repo_key = str(repo_path)
            self.watched_repos[repo_key] = watched_repo
            
            logger.info(
                f"Registered repository: {repo_path} "
                f"(remote: {remote_url}, branch: {branch})"
            )
            return True
            
        except Exception as e:
            logger.error(f"Error registering repository {repo_path}: {e}")
            return False
    
    def _unregister_repository(self, repo_path: Path):
        """
        Unregister a repository.
        
        Args:
            repo_path: Path to the repository
        """
        repo_key = str(repo_path)
        if repo_key in self.watched_repos:
            del self.watched_repos[repo_key]
            logger.info(f"Unregistered repository: {repo_path}")
    
    def _scan_directory(self, directory: Path):
        """
        Recursively scan directory for repositories.
        
        Args:
            directory: Directory to scan
        """
        try:
            if not directory.exists() or not directory.is_dir():
                return
            
            # Check if current directory is a repository
            if (directory / ".git").exists():
                self._register_repository(directory)
                # Don't scan subdirectories of a git repo
                return
            
            # Recursively scan subdirectories
            for item in directory.iterdir():
                if item.is_dir() and not item.name.startswith('.'):
                    self._scan_directory(item)
                    
        except PermissionError:
            logger.warning(f"Permission denied accessing {directory}")
        except Exception as e:
            logger.error(f"Error scanning directory {directory}: {e}")
    
    def initial_scan(self):
        """Perform initial scan of the root path."""
        logger.info(f"Starting initial scan of {self.root_path}")
        start_time = time.time()
        
        self._scan_directory(self.root_path)
        
        elapsed = time.time() - start_time
        logger.info(
            f"Initial scan complete. Found {len(self.watched_repos)} repositories "
            f"in {elapsed:.2f} seconds"
        )
    
    def start_monitoring(self):
        """Start monitoring for filesystem changes."""
        if self.observer:
            logger.warning("Monitor already running")
            return
        
        # Create event handler
        event_handler = RepositoryEventHandler(
            on_config_added=self._register_repository,
            on_config_removed=self._unregister_repository
        )
        
        # Create and start observer
        self.observer = Observer()
        self.observer.schedule(event_handler, str(self.root_path), recursive=True)
        self.observer.start()
        
        logger.info(f"Started monitoring {self.root_path}")
    
    def stop_monitoring(self):
        """Stop monitoring for filesystem changes."""
        if self.observer:
            self.observer.stop()
            self.observer.join()
            self.observer = None
            logger.info("Stopped monitoring")
    
    def find_repository_by_remote(self, remote_url: str, branch: str) -> Optional[WatchedRepository]:
        """
        Find a watched repository by remote URL and branch.
        
        Args:
            remote_url: Git remote URL
            branch: Branch name
            
        Returns:
            WatchedRepository if found, None otherwise
        """
        for repo in self.watched_repos.values():
            try:
                git_manager = GitManager(Path(repo.path))
                if git_manager.matches_remote(remote_url) and repo.branch == branch:
                    return repo
            except Exception as e:
                logger.error(f"Error checking repository {repo.path}: {e}")
                continue
        
        return None
    
    def get_all_repositories(self) -> Dict[str, WatchedRepository]:
        """
        Get all watched repositories.
        
        Returns:
            Dictionary of watched repositories
        """
        return self.watched_repos.copy()
    
    def get_repository_count(self) -> int:
        """
        Get the number of watched repositories.
        
        Returns:
            Number of repositories
        """
        return len(self.watched_repos)
