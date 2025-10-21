"""Main application entry point."""

import logging
import os
import sys
from pathlib import Path
from typing import Optional
import uvicorn
from pydantic_settings import BaseSettings

from .folder_monitor import FolderMonitor
from .email_notifier import EmailNotifier
from .webhook_server import WebhookServer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler("/app/logs/app.log"),
    ],
)

logger = logging.getLogger(__name__)


class Settings(BaseSettings):
    """Application settings from environment variables."""

    # GitHub App settings
    github_app_id: str
    github_app_private_key_path: str
    github_webhook_secret: str

    # SMTP settings
    smtp_host: str
    smtp_port: int = 587
    smtp_user: str
    smtp_password: str
    smtp_from_email: str
    default_notification_email: str

    # Application settings
    webhook_port: int = 8080
    repos_path: str = "/repos"
    log_level: str = "INFO"

    class Config:
        env_file = ".env"
        case_sensitive = False


def setup_logging(log_level: str):
    """
    Setup logging configuration.

    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR)
    """
    level = getattr(logging, log_level.upper(), logging.INFO)
    logging.getLogger().setLevel(level)

    # Set level for all loggers
    for logger_name in ["src", "uvicorn", "fastapi"]:
        logging.getLogger(logger_name).setLevel(level)


def validate_settings(settings: Settings) -> bool:
    """
    Validate application settings.

    Args:
        settings: Settings instance

    Returns:
        True if valid, False otherwise
    """
    errors = []

    # Check repos path exists
    repos_path = Path(settings.repos_path)
    if not repos_path.exists():
        errors.append(f"Repos path does not exist: {settings.repos_path}")

    # Check private key file exists
    private_key_path = Path(settings.github_app_private_key_path)
    if not private_key_path.exists():
        errors.append(
            f"Private key file does not exist: {settings.github_app_private_key_path}"
        )
    elif not private_key_path.is_file():
        errors.append(
            f"Private key path is not a file: {settings.github_app_private_key_path}"
        )

    # Check required settings
    required_fields = [
        "github_app_id",
        "github_webhook_secret",
        "smtp_host",
        "smtp_user",
        "smtp_password",
        "smtp_from_email",
        "default_notification_email",
    ]

    for field in required_fields:
        value = getattr(settings, field, None)
        if not value:
            errors.append(f"Required setting missing: {field.upper()}")

    if errors:
        for error in errors:
            logger.error(error)
        return False

    return True


def create_log_directory():
    """Create log directory if it doesn't exist."""
    log_dir = Path("/app/logs")
    log_dir.mkdir(parents=True, exist_ok=True)


def main():
    """Main application entry point."""
    logger.info("Starting GitHub Auto-Deployer")

    # Create log directory
    create_log_directory()

    # Load settings
    try:
        settings = Settings()
    except Exception as e:
        logger.error(f"Failed to load settings: {e}")
        sys.exit(1)

    # Setup logging
    setup_logging(settings.log_level)

    # Validate settings
    if not validate_settings(settings):
        logger.error("Invalid configuration. Exiting.")
        sys.exit(1)

    logger.info(f"Configuration loaded successfully")
    logger.info(f"Repos path: {settings.repos_path}")
    logger.info(f"Webhook port: {settings.webhook_port}")
    logger.info(f"Default notification email: {settings.default_notification_email}")

    # Initialize components
    try:
        # Initialize email notifier
        email_notifier = EmailNotifier(
            smtp_host=settings.smtp_host,
            smtp_port=settings.smtp_port,
            smtp_user=settings.smtp_user,
            smtp_password=settings.smtp_password,
            from_email=settings.smtp_from_email,
        )
        logger.info("Email notifier initialized")

        # Initialize folder monitor
        folder_monitor = FolderMonitor(
            root_path=Path(settings.repos_path),
            default_email=settings.default_notification_email,
        )
        logger.info("Folder monitor initialized")

        # Perform initial scan
        folder_monitor.initial_scan()

        # Start monitoring
        folder_monitor.start_monitoring()
        logger.info("Folder monitoring started")

        # Initialize webhook server
        webhook_server = WebhookServer(
            folder_monitor=folder_monitor,
            email_notifier=email_notifier,
            webhook_secret=settings.github_webhook_secret,
        )
        logger.info("Webhook server initialized")

        # Get FastAPI app
        app = webhook_server.get_app()

        # Start server
        logger.info(f"Starting webhook server on port {settings.webhook_port}")
        uvicorn.run(
            app,
            host="0.0.0.0",
            port=settings.webhook_port,
            log_level=settings.log_level.lower(),
        )

    except KeyboardInterrupt:
        logger.info("Received shutdown signal")
    except Exception as e:
        logger.error(f"Fatal error: {e}", exc_info=True)
        sys.exit(1)
    finally:
        # Cleanup
        if "folder_monitor" in locals():
            folder_monitor.stop_monitoring()
            logger.info("Folder monitoring stopped")

        logger.info("GitHub Auto-Deployer stopped")


if __name__ == "__main__":
    main()
