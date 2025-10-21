# GitHub App Setup Guide

This guide will walk you through creating and configuring a GitHub App for the GitHub Auto-Deployer.

## Overview

A GitHub App is required to receive webhook notifications when you push changes to your repositories. The app will send POST requests to your deployer instance whenever a push event occurs.

## Step 1: Create a GitHub App

1. **Navigate to GitHub App Settings**

   - Go to your GitHub account settings
   - Click on "Developer settings" in the left sidebar
   - Click on "GitHub Apps"
   - Click "New GitHub App"

2. **Basic Information**

   Fill in the following fields:

   - **GitHub App name**: Choose a unique name (e.g., "My Auto Deployer" or "Production Deployer")
   - **Homepage URL**: Your server URL or the repository URL of this project
   - **Webhook URL**: `https://your-domain.com/webhook`
     - This must be publicly accessible
     - Must use HTTPS (GitHub requirement)
     - Example: `https://deploy.example.com/webhook`
   - **Webhook secret**: Generate a strong random secret
     - Use a password generator or run: `openssl rand -hex 32`
     - Save this value - you'll need it for your `.env` file

3. **Permissions**

   Set the following repository permissions:

   - **Contents**: Read-only
     - This allows the app to read repository content
     - Required for the webhook to include repository information

4. **Subscribe to Events**

   Check the following event:

   - ✅ **Push** - Triggered when commits are pushed to a repository

5. **Where can this GitHub App be installed?**

   Choose based on your needs:

   - **Only on this account** - If you only need it for your personal repositories
   - **Any account** - If you want to share it with others (not recommended for personal deployments)

6. **Create the GitHub App**

   Click "Create GitHub App" at the bottom of the page.

## Step 2: Generate and Configure Private Key

After creating the app, you need to generate a private key:

1. **Generate Private Key**

   - On your GitHub App's settings page, scroll down to "Private keys"
   - Click "Generate a private key"
   - A `.pem` file will be downloaded to your computer

2. **Save the Private Key File**

   Save the downloaded `.pem` file to a secure location on your server. For example:

   ```bash
   # Create a secure directory for the key
   mkdir -p /etc/github-auto-deployer

   # Move the key file
   mv ~/Downloads/your-private-key.pem /etc/github-auto-deployer/github-app-key.pem

   # Set secure permissions (read-only for owner)
   chmod 600 /etc/github-auto-deployer/github-app-key.pem
   ```

3. **Configure the Path**

   Set the path to the private key file in your `.env` file:

   ```env
   GITHUB_APP_PRIVATE_KEY_PATH=/app/github-app-key.pem
   ```

   Note: The path `/app/github-app-key.pem` is the path inside the Docker container. You'll mount your actual file to this location in `docker-compose.yml`.

4. **Update docker-compose.yml**

   Add a volume mount for the private key file:

   ```yaml
   volumes:
     - /etc/github-auto-deployer/github-app-key.pem:/app/github-app-key.pem:ro
   ```

   The `:ro` flag makes it read-only for extra security.

## Step 3: Note Your App ID

On your GitHub App's settings page, you'll see the **App ID** near the top. Copy this value to your `.env` file:

```env
GITHUB_APP_ID=123456
```

## Step 4: Configure Webhook Secret

Copy the webhook secret you generated in Step 1 to your `.env` file:

```env
GITHUB_WEBHOOK_SECRET=your_webhook_secret_here
```

## Step 5: Install the GitHub App

Now you need to install the app on the repositories you want to deploy:

1. **Navigate to Installation**

   - On your GitHub App's settings page
   - Click "Install App" in the left sidebar
   - Or go to: `https://github.com/apps/your-app-name/installations/new`

2. **Choose Account**

   - Select your personal account or organization

3. **Select Repositories**

   - Choose "All repositories" or "Only select repositories"
   - If selecting specific repositories, choose the ones you want to auto-deploy

4. **Install**
   - Click "Install"
   - You'll be redirected to a confirmation page

## Step 6: Verify Installation

1. **Check Installed Apps**

   - Go to your repository settings
   - Click "Integrations" in the left sidebar
   - You should see your GitHub App listed

2. **Test Webhook Delivery**
   - Make sure your deployer is running
   - Push a commit to a watched repository
   - Go to your GitHub App settings → Advanced → Recent Deliveries
   - You should see webhook deliveries with 200 status codes

## Complete .env Configuration

Your final `.env` file should look like this:

```env
# GitHub App Configuration
GITHUB_APP_ID=123456
GITHUB_APP_PRIVATE_KEY_PATH=/app/github-app-key.pem
GITHUB_WEBHOOK_SECRET=your_webhook_secret_here

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_EMAIL=your_email@gmail.com

# Notification Configuration
DEFAULT_NOTIFICATION_EMAIL=admin@example.com
```

And ensure your `docker-compose.yml` has the volume mount:

```yaml
volumes:
  - /path/to/your/repos:/repos:rw
  - ./logs:/app/logs
  - /etc/github-auto-deployer/github-app-key.pem:/app/github-app-key.pem:ro
```

## Troubleshooting

### Webhook Deliveries Failing

**Problem**: Webhooks show failed deliveries in GitHub

**Solutions**:

1. Ensure your webhook URL is publicly accessible
2. Verify you're using HTTPS (required by GitHub)
3. Check that your deployer is running: `docker compose ps`
4. Check deployer logs: `docker compose logs -f`
5. Verify the webhook secret matches in both GitHub and `.env`

### Invalid Signature Errors

**Problem**: Logs show "Invalid webhook signature"

**Solutions**:

1. Verify the webhook secret in `.env` matches exactly what's in GitHub
2. Ensure there are no extra spaces or quotes in the `.env` file
3. Restart the container after changing `.env`: `docker compose restart`

### Private Key Errors

**Problem**: Application fails to start with private key errors

**Solutions**:

1. Verify the private key file exists at the specified path
2. Check file permissions: `ls -l /etc/github-auto-deployer/github-app-key.pem`
3. Ensure the file is mounted correctly in `docker-compose.yml`
4. Verify the path in `.env` matches the container path: `/app/github-app-key.pem`
5. Check the private key file format (should be a valid PEM file)

### Repository Not Found

**Problem**: Webhook received but repository not found

**Solutions**:

1. Ensure the repository has a `.deployer.yml` file
2. Verify the repository is in the mounted volume path
3. Check that the repository has a `.git` directory
4. Review logs to see what repositories were detected on startup

## Security Best Practices

1. **Webhook Secret**

   - Use a strong, random secret (at least 32 characters)
   - Never commit the secret to version control
   - Rotate the secret periodically

2. **Private Key**

   - Store the `.pem` file in a secure location outside the project directory
   - Never commit the private key to version control
   - Restrict file permissions: `chmod 600 /path/to/github-app-key.pem`
   - Consider using a secrets management system for production deployments

3. **HTTPS**

   - Always use HTTPS for the webhook endpoint
   - Use a valid SSL certificate (Let's Encrypt is free)
   - Consider using a reverse proxy (nginx, Caddy, Traefik)

4. **Access Control**
   - Only install the app on repositories that need auto-deployment
   - Regularly review installed repositories
   - Remove the app from repositories that no longer need it

## Testing Your Setup

1. **Start the Deployer**

   ```bash
   docker compose up -d
   docker compose logs -f
   ```

2. **Check Repository Detection**

   ```bash
   curl http://localhost:8080/repositories
   ```

3. **Make a Test Push**

   - Push a commit to a watched repository
   - Check the deployer logs for webhook processing
   - Verify the deployment command was executed

4. **Check GitHub Webhook Deliveries**
   - Go to your GitHub App settings
   - Click "Advanced" → "Recent Deliveries"
   - Click on a delivery to see the request and response

## Additional Resources

- [GitHub Apps Documentation](https://docs.github.com/en/developers/apps)
- [Webhooks Documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [GitHub App Permissions](https://docs.github.com/en/developers/apps/building-github-apps/setting-permissions-for-github-apps)

## Need Help?

If you encounter issues not covered in this guide:

1. Check the deployer logs: `docker compose logs -f`
2. Review GitHub webhook delivery details
3. Verify all configuration values in `.env`
4. Open an issue on the project repository with:
   - Error messages from logs
   - GitHub webhook delivery details (remove sensitive data)
   - Your configuration (remove secrets)
