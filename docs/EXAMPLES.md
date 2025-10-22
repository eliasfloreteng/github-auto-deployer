# Usage Examples

This document provides practical examples of using the GitHub Auto Deployer in various scenarios.

## Example 1: Docker Compose Application

### Scenario

You have a web application that runs with Docker Compose and want to automatically redeploy when you push to the `main` branch.

### Setup

1. **Clone your repository on the server:**

   ```bash
   cd /opt
   git clone https://github.com/username/myapp.git
   cd myapp
   git checkout main
   ```

2. **Add the folder to the deployer:**

   ```bash
   ./deployer add-folder
   ```

   When prompted:

   - Repository Path: `/opt/myapp`
   - Command: `docker compose up -d --pull=auto --build`

3. **Test it:**

   ```bash
   # Make a change and push
   echo "# Update" >> README.md
   git add README.md
   git commit -m "Test deployment"
   git push

   # Watch the logs
   sudo journalctl -u github-deployer -f
   ```

### Expected Behavior

- Webhook received from GitHub
- Git pull executed
- Docker Compose rebuilds and restarts containers
- Application is updated with zero downtime (if configured properly)

## Example 2: Node.js Application with PM2

### Scenario

You have a Node.js application managed by PM2 that needs to be rebuilt and restarted on each deployment.

### Setup

1. **Clone and set up:**

   ```bash
   cd /var/www
   git clone https://github.com/username/node-app.git
   cd node-app
   npm install
   pm2 start ecosystem.config.js
   pm2 save
   ```

2. **Add to deployer:**

   ```bash
   ./deployer add-folder
   ```

   Command: `npm install && npm run build && pm2 restart node-app`

### Configuration File Example

```json
{
  "folders": [
    {
      "path": "/var/www/node-app",
      "command": "npm install && npm run build && pm2 restart node-app",
      "branch": "production",
      "repo_url": "https://github.com/username/node-app"
    }
  ]
}
```

## Example 3: Static Website with Hugo

### Scenario

You have a Hugo static site that needs to be rebuilt and deployed to nginx.

### Setup

1. **Clone repository:**

   ```bash
   cd /home/user/sites
   git clone https://github.com/username/blog.git
   cd blog
   ```

2. **Add to deployer:**

   ```bash
   ./deployer add-folder
   ```

   Command: `hugo --minify && rsync -av --delete public/ /var/www/html/blog/`

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name blog.example.com;
    root /var/www/html/blog;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

## Example 4: Multiple Environments

### Scenario

You want to deploy different branches to different environments (staging and production).

### Setup

1. **Staging environment:**

   ```bash
   cd /opt
   git clone https://github.com/username/app.git app-staging
   cd app-staging
   git checkout staging

   ./deployer add-folder
   # Path: /opt/app-staging
   # Command: docker compose -f docker-compose.staging.yml up -d --build
   ```

2. **Production environment:**

   ```bash
   cd /opt
   git clone https://github.com/username/app.git app-production
   cd app-production
   git checkout main

   ./deployer add-folder
   # Path: /opt/app-production
   # Command: docker compose -f docker-compose.prod.yml up -d --build
   ```

### Result

- Pushes to `staging` branch → deploys to staging environment
- Pushes to `main` branch → deploys to production environment

## Example 5: Python Application with systemd

### Scenario

You have a Python application running as a systemd service.

### Setup

1. **Clone and set up:**

   ```bash
   cd /opt
   git clone https://github.com/username/python-app.git
   cd python-app
   python3 -m venv venv
   source venv/bin/activate
   pip install -r requirements.txt
   ```

2. **Create systemd service:**

   ```ini
   # /etc/systemd/system/python-app.service
   [Unit]
   Description=Python Application
   After=network.target

   [Service]
   Type=simple
   User=www-data
   WorkingDirectory=/opt/python-app
   Environment="PATH=/opt/python-app/venv/bin"
   ExecStart=/opt/python-app/venv/bin/python app.py
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

3. **Add to deployer:**

   ```bash
   ./deployer add-folder
   ```

   Command: `source venv/bin/activate && pip install -r requirements.txt && sudo systemctl restart python-app`

## Example 6: Database Migrations

### Scenario

You need to run database migrations before restarting the application.

### Setup

```bash
./deployer add-folder
```

Command: `docker compose run --rm app python manage.py migrate && docker compose up -d --build`

### Alternative with Separate Migration Service

```bash
./deployer add-folder
```

Command: `./scripts/deploy.sh`

**deploy.sh:**

```bash
#!/bin/bash
set -e

echo "Running migrations..."
docker compose run --rm app python manage.py migrate

echo "Collecting static files..."
docker compose run --rm app python manage.py collectstatic --noinput

echo "Restarting services..."
docker compose up -d --build

echo "Deployment complete!"
```

## Example 7: Rollback on Failure

### Scenario

You want to automatically rollback if the deployment command fails.

### Setup

Create a deployment script with rollback logic:

**deploy-with-rollback.sh:**

```bash
#!/bin/bash
set -e

# Save current commit
CURRENT_COMMIT=$(git rev-parse HEAD)

# Try to deploy
if docker compose up -d --build; then
    echo "Deployment successful"
    exit 0
else
    echo "Deployment failed, rolling back..."
    git reset --hard $CURRENT_COMMIT
    docker compose up -d --build
    exit 1
fi
```

Add to deployer:

```bash
./deployer add-folder
# Command: ./deploy-with-rollback.sh
```

## Example 8: Notification on Success

### Scenario

You want to send a notification when deployment succeeds (not just on failure).

### Setup

Create a wrapper script:

**deploy-with-notification.sh:**

```bash
#!/bin/bash

# Run deployment
if docker compose up -d --build; then
    # Send success notification (using curl to a webhook)
    curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
         -H 'Content-Type: application/json' \
         -d '{"text":"Deployment successful for myapp"}'
    exit 0
else
    # Failure will be handled by email notification
    exit 1
fi
```

## Example 9: Conditional Deployment

### Scenario

Only deploy if tests pass.

### Setup

**deploy-with-tests.sh:**

```bash
#!/bin/bash
set -e

echo "Running tests..."
docker compose run --rm app npm test

if [ $? -eq 0 ]; then
    echo "Tests passed, deploying..."
    docker compose up -d --build
else
    echo "Tests failed, aborting deployment"
    exit 1
fi
```

## Example 10: Multi-Service Deployment

### Scenario

You have multiple services that need to be deployed in a specific order.

### Setup

**deploy-multi-service.sh:**

```bash
#!/bin/bash
set -e

echo "Deploying database migrations..."
docker compose run --rm db-migrate

echo "Deploying backend..."
docker compose up -d --build backend

echo "Waiting for backend to be ready..."
sleep 10

echo "Deploying frontend..."
docker compose up -d --build frontend

echo "Deploying worker..."
docker compose up -d --build worker

echo "All services deployed successfully!"
```

## Troubleshooting Examples

### Check if webhook is being received

```bash
sudo journalctl -u github-deployer -f | grep "Processing push event"
```

### Test git pull manually

```bash
cd /path/to/repo
sudo -u www-data git pull
```

### Test command execution manually

```bash
cd /path/to/repo
sudo -u www-data docker compose up -d --build
```

### View full deployment logs

```bash
sudo journalctl -u github-deployer --since "1 hour ago" --no-pager
```

## Best Practices

1. **Always test commands manually first** before adding them to the deployer
2. **Use absolute paths** in deployment scripts
3. **Set appropriate timeouts** for long-running commands
4. **Include health checks** in your deployment scripts
5. **Keep deployment scripts idempotent** (safe to run multiple times)
6. **Log everything** for debugging purposes
7. **Use environment-specific configurations** (staging vs production)
8. **Test rollback procedures** before you need them
9. **Monitor disk space** to prevent failed deployments
10. **Keep secrets out of git** (use environment variables or secret management)
