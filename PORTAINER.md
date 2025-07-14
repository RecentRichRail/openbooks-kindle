# OpenBooks Portainer Deployment Guide

This guide walks you through deploying OpenBooks with Send to Kindle functionality using Portainer.

## 🚀 Quick Deploy in Portainer

### Method 1: Git Repository Deploy (Recommended)

1. **In Portainer, go to Stacks → Add Stack**
2. **Choose "Repository" as build method**
3. **Enter your repository details:**
   - Repository URL: `https://github.com/RecentRichRail/openbooks-kindle`
   - Reference: `master`
   - Compose path: `docker-compose.yml`

4. **Configure Environment Variables in Portainer:**
   ```
   SMTP_ENABLED=true
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your_email@gmail.com
   SMTP_PASSWORD=your_app_password
   SMTP_FROM=your_email@gmail.com
   ```

5. **Deploy the stack**

### Method 2: Web Editor Deploy

1. **In Portainer, go to Stacks → Add Stack**
2. **Choose "Web editor"**
3. **Copy and paste your docker-compose.yml content**
4. **Add environment variables as above**
5. **Deploy the stack**

## 📧 SMTP Configuration for Portainer

### Environment Variables Setup

In Portainer's stack configuration, add these environment variables:

#### For Gmail:
```
SMTP_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com
```

#### For iCloud Mail:
```
SMTP_ENABLED=true
SMTP_HOST=smtp.mail.me.com
SMTP_PORT=587
SMTP_USERNAME=your_email@me.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@me.com
```

#### For Outlook/Hotmail:
```
SMTP_ENABLED=true
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587
SMTP_USERNAME=your_email@outlook.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@outlook.com
```

### Important Notes:
- **Use App Passwords**: Never use your regular email password
- **Enable 2FA**: Required for app passwords on most providers
- **Test Configuration**: Use a test email first

## 🔧 Portainer-Specific Configuration

### Stack Settings Recommendations:

1. **Stack Name**: `openbooks-kindle`
2. **Access Control**: Set appropriate user/team access
3. **Auto-update**: Enable if you want automatic updates from git
4. **Webhook**: Optional - for CI/CD integration

### Volume Management:

Your stack creates these volumes:
- `openbooks-kindle_books_data` - Downloaded books storage
- `openbooks-kindle_logs` - Application logs (if logs directory exists)

### Port Configuration:

- **Container Port**: 80
- **Host Port**: 5228 (as configured)
- **Access URL**: `http://your-server:5228`

## 🔍 Monitoring in Portainer

### Container Health:
- Check the health status in Portainer's container view
- Green = healthy, Red = unhealthy
- Health check runs every 30 seconds

### Logs Access:
1. Go to **Containers** → **openbooks-kindle**
2. Click **Logs** to view real-time application logs
3. Look for:
   - `SMTP configuration loaded successfully` - SMTP working
   - `Connected to IRC` - IRC connection established
   - `Server started on port 80` - Server running

### Common Log Messages:
```
✅ Good:
- "SMTP configuration loaded successfully"
- "Connected to IRC Highway"
- "Server started on port 80"

❌ Issues:
- "SMTP configuration failed"
- "Failed to connect to IRC"
- "Authentication failed"
```

## 🛠️ Troubleshooting in Portainer

### Build Errors

#### "unrecognized token '<'" Error

This error typically occurs when:

1. **Git LFS Files**: Binary files are replaced with HTML pointers
2. **Network Issues**: Files downloaded as HTML error pages
3. **Corrupted Files**: Binary files contain HTML content

**Solutions:**

1. **Clean Build Environment:**
   ```bash
   # Run the pre-build script (if available locally)
   ./portainer-build.sh
   ```

2. **Force Rebuild in Portainer:**
   - Go to **Stacks** → **openbooks-kindle**
   - Click **Editor**
   - Check "Re-pull image and rebuild"
   - Click **Update the stack**

3. **Check Repository Access:**
   - Ensure Portainer can access your GitHub repository
   - Verify repository permissions
   - Try using HTTPS instead of SSH repository URL

4. **Verify .dockerignore:**
   - Make sure `.dockerignore` excludes problematic files:
     ```
     **/*.epub
     **/*.zip
     cmd/mock_server/great-gatsby.epub
     cmd/mock_server/SearchBot_results_for__the_great_gatsby.txt.zip
     ```

5. **Alternative: Use Pre-built Image:**
   - Build the image locally first
   - Push to Docker Hub or registry
   - Use the pre-built image in Portainer

#### Frontend Build Failures:

```
Error: npm install failed
```

**Solution:**
- Clear npm cache in build
- Use `npm ci` instead of `npm install`
- Check Node.js version compatibility

### SMTP Issues:

1. **Check Environment Variables:**
   - Go to **Stacks** → **openbooks-kindle** → **Editor**
   - Verify all SMTP variables are set correctly

2. **View Container Logs:**
   ```
   Look for: "SMTP Error" or "Authentication failed"
   ```

3. **Test Email Provider Settings:**
   - Verify SMTP host and port
   - Confirm app password is correct
   - Check if 2FA is enabled

### Container Won't Start:

1. **Check Build Logs:**
   - Go to **Stacks** → **openbooks-kindle**
   - Look for build errors in the deployment logs

2. **Verify Resources:**
   - Ensure sufficient CPU/memory
   - Check disk space for volumes

3. **Port Conflicts:**
   - Verify port 5228 isn't used by another service
   - Change port in compose file if needed

### Portainer-Specific Issues:

#### Repository Access Problems:
1. **Check Repository URL:**
   - Use HTTPS: `https://github.com/RecentRichRail/openbooks-kindle.git`
   - Avoid SSH URLs in Portainer
   - Ensure repository is public or provide credentials

2. **Branch Issues:**
   - Verify branch name (`master` vs `main`)
   - Check if branch exists
   - Try using specific commit hash

3. **Build Context Problems:**
   - Dockerfile must be in repository root
   - docker-compose.yml must be in repository root
   - Check file permissions

4. **Resource Limitations:**
   - Increase build timeout in Portainer
   - Check available disk space
   - Monitor memory usage during build

#### Stack Deployment Issues:

1. **Environment Variables Not Loading:**
   ```bash
   # Check in container
   docker exec openbooks-kindle env | grep SMTP
   ```

2. **Volume Mount Issues:**
   ```bash
   # Verify volumes exist
   docker volume ls | grep openbooks
   
   # Check volume permissions
   docker exec openbooks-kindle ls -la /books
   ```

3. **Port Conflicts:**
   ```bash
   # Check if port 5228 is in use
   netstat -tulpn | grep :5228
   ```

### Health Check Failures:

1. **Container Status:**
   - Red status = health check failing
   - Check if wget is available in container

2. **Network Issues:**
   - Verify container can reach localhost:80
   - Check firewall settings

## 📱 Accessing Your Application

Once deployed successfully:

1. **Web Interface**: `http://your-server-ip:5228`
2. **Search for Books**: Use the search functionality
3. **Send to Kindle**: Click "Request" button on any book
4. **Enter Email**: Use your Kindle email address
5. **Check Email**: Book should arrive in your Kindle library

## 🔄 Updates and Maintenance

### Updating the Application:

1. **Git Repository Method:**
   - Portainer will auto-pull updates if configured
   - Or manually redeploy the stack

2. **Manual Update:**
   - Go to **Stacks** → **openbooks-kindle**
   - Click **Editor** → **Update the stack**

### Backup Considerations:

- **Books Volume**: `books_data` contains downloaded files
- **Environment Variables**: Export stack configuration
- **Logs**: Optional - for troubleshooting history

## 🚨 Security Best Practices

1. **Environment Variables:**
   - Never hardcode passwords in compose file
   - Use Portainer's environment variable system
   - Consider using Docker secrets for production

2. **Network Security:**
   - Use reverse proxy (Traefik/Nginx) if exposing to internet
   - Enable HTTPS with SSL certificates
   - Restrict access by IP if needed

3. **Container Security:**
   - Keep base images updated
   - Run with non-root user (already configured)
   - Monitor security advisories

## 📞 Support

If you encounter issues:

1. **Check Portainer Logs**: Most detailed error information
2. **Verify SMTP Settings**: Common cause of issues
3. **Test IRC Connection**: Ensure network connectivity
4. **Resource Monitoring**: Check CPU/memory usage

Remember: The application includes random IRC username generation, enhanced notification detection, and proper Send to Kindle functionality - all working seamlessly in your Portainer deployment!
