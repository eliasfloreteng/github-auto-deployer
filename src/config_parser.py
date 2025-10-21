"""Configuration parser for .deployer.yml files."""
import logging
from pathlib import Path
from typing import Optional
import yaml
from .models import DeployerConfig

logger = logging.getLogger(__name__)


class ConfigParser:
    """Parser for .deployer.yml configuration files."""
    
    CONFIG_FILENAME = ".deployer.yml"
    
    @classmethod
    def find_config(cls, directory: Path) -> Optional[Path]:
        """
        Find .deployer.yml file in the given directory.
        
        Args:
            directory: Directory to search in
            
        Returns:
            Path to config file if found, None otherwise
        """
        config_path = directory / cls.CONFIG_FILENAME
        if config_path.exists() and config_path.is_file():
            return config_path
        return None
    
    @classmethod
    def parse_config(cls, config_path: Path) -> Optional[DeployerConfig]:
        """
        Parse a .deployer.yml configuration file.
        
        Args:
            config_path: Path to the configuration file
            
        Returns:
            DeployerConfig object if valid, None otherwise
        """
        try:
            with open(config_path, 'r') as f:
                data = yaml.safe_load(f)
            
            if not data:
                logger.error(f"Empty configuration file: {config_path}")
                return None
            
            # Validate required fields
            if 'command' not in data:
                logger.error(f"Missing required 'command' field in {config_path}")
                return None
            
            config = DeployerConfig(**data)
            logger.info(f"Successfully parsed config from {config_path}")
            return config
            
        except yaml.YAMLError as e:
            logger.error(f"YAML parsing error in {config_path}: {e}")
            return None
        except Exception as e:
            logger.error(f"Error parsing config {config_path}: {e}")
            return None
    
    @classmethod
    def is_valid_repository(cls, directory: Path) -> bool:
        """
        Check if directory is a valid git repository with a config file.
        
        Args:
            directory: Directory to check
            
        Returns:
            True if valid repository with config, False otherwise
        """
        if not directory.is_dir():
            return False
        
        # Check for .git directory
        git_dir = directory / ".git"
        if not git_dir.exists():
            return False
        
        # Check for config file
        config_path = cls.find_config(directory)
        if not config_path:
            return False
        
        # Try to parse config
        config = cls.parse_config(config_path)
        return config is not None
