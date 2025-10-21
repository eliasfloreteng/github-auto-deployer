"""Command executor with timeout support."""
import logging
import subprocess
import shlex
from pathlib import Path
from typing import Tuple
import signal
import os

logger = logging.getLogger(__name__)


class CommandExecutor:
    """Executes shell commands with timeout support."""
    
    def __init__(self, working_directory: Path):
        """
        Initialize CommandExecutor.
        
        Args:
            working_directory: Directory to execute commands in
        """
        self.working_directory = working_directory
    
    def execute(self, command: str, timeout: int = 300) -> Tuple[bool, str, int]:
        """
        Execute a shell command with timeout.
        
        Args:
            command: Command to execute
            timeout: Timeout in seconds
            
        Returns:
            Tuple of (success, output, exit_code)
        """
        logger.info(f"Executing command in {self.working_directory}: {command}")
        logger.info(f"Timeout: {timeout} seconds")
        
        try:
            # Execute command in shell to support complex commands like docker compose
            process = subprocess.Popen(
                command,
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                cwd=str(self.working_directory),
                text=True,
                preexec_fn=os.setsid if os.name != 'nt' else None
            )
            
            try:
                output, _ = process.communicate(timeout=timeout)
                exit_code = process.returncode
                
                if exit_code == 0:
                    logger.info(f"Command completed successfully with exit code {exit_code}")
                    return True, output, exit_code
                else:
                    logger.error(f"Command failed with exit code {exit_code}")
                    return False, output, exit_code
                    
            except subprocess.TimeoutExpired:
                logger.error(f"Command timed out after {timeout} seconds")
                
                # Kill the process group
                if os.name != 'nt':
                    os.killpg(os.getpgid(process.pid), signal.SIGTERM)
                else:
                    process.terminate()
                
                # Wait a bit for graceful termination
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    # Force kill if still running
                    if os.name != 'nt':
                        os.killpg(os.getpgid(process.pid), signal.SIGKILL)
                    else:
                        process.kill()
                
                output = f"Command timed out after {timeout} seconds and was terminated"
                return False, output, -1
                
        except Exception as e:
            error_msg = f"Error executing command: {str(e)}"
            logger.error(error_msg)
            return False, error_msg, -1
    
    def validate_command(self, command: str) -> Tuple[bool, str]:
        """
        Validate a command before execution.
        
        Args:
            command: Command to validate
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        if not command or not command.strip():
            return False, "Command is empty"
        
        # Check for potentially dangerous commands
        dangerous_patterns = [
            'rm -rf /',
            'mkfs',
            'dd if=',
            '> /dev/sda',
        ]
        
        command_lower = command.lower()
        for pattern in dangerous_patterns:
            if pattern in command_lower:
                return False, f"Command contains potentially dangerous pattern: {pattern}"
        
        return True, ""
    
    def execute_safe(self, command: str, timeout: int = 300) -> Tuple[bool, str, int]:
        """
        Execute a command after validation.
        
        Args:
            command: Command to execute
            timeout: Timeout in seconds
            
        Returns:
            Tuple of (success, output, exit_code)
        """
        # Validate command first
        is_valid, error_msg = self.validate_command(command)
        if not is_valid:
            logger.error(f"Command validation failed: {error_msg}")
            return False, f"Command validation failed: {error_msg}", -1
        
        # Execute the command
        return self.execute(command, timeout)
