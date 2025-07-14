# OpenBooks Docker Deployment

> See [Github](https://github.com/evan-buss/openbooks) for more information.

## Quick Start with Send to Kindle

### 1. Configure Environment Variables

Create a `.env` file from the example:
```bash
cp .env.example .env
```

Edit `.env` with your SMTP settings:
```bash
SMTP_ENABLED=true
SMTP_HOST=smtp.mail.me.com  # For iCloud Mail
SMTP_PORT=587
SMTP_USERNAME=your_email@me.com
SMTP_PASSWORD=your_app_password  # Use App Password
SMTP_FROM=your_email@me.com
```

### 2. Start with Docker Compose

```bash
docker-compose up -d
```

Access at: http://localhost:8080

## Supported Email Providers

| Provider | SMTP Host | Port | Notes |
|----------|-----------|------|-------|
| iCloud Mail | smtp.mail.me.com | 587 | Use App Password |
| Gmail | smtp.gmail.com | 587 | Use App Password |
| Outlook/Hotmail | smtp-mail.outlook.com | 587 | Use App Password |
| Yahoo | smtp.mail.yahoo.com | 587 | Use App Password |

## Basic Usage (Without Email)

### Simple Run
`docker run -d -p 8080:80 evanbuss/openbooks --name my_irc_name`

### Persist eBook Files
`docker run -d -p 8080:80 -v ~/Downloads:/books evanbuss/openbooks --name my_irc_name --persist`

### Host at Sub Path
`docker run -d -p 8080:80 -e BASE_PATH=/openbooks/ evanbuss/openbooks --name my_irc_name`

## Arguments

```
--name string
    Required name when connecting to irchighway (auto-generated if not provided)
--persist
    Keep book files in the download dir. Default is to delete after sending.
--no-browser-downloads
    Disable direct browser downloads (recommended with --persist)
--port
    Server port (default: 80 in container, 8080 on host)
--log-level
    Logging level: debug, info, warn, error
```

## Docker Compose (Recommended)

```yaml
version: '3.8'
services:
  openbooks:
    build: .
    ports:
      - '8080:8080'
    volumes:
      - ./data:/app/data
      - ./.env:/app/.env
    restart: unless-stopped
    container_name: openbooks
    environment:
      # SMTP settings loaded from .env file
      - SMTP_ENABLED=${SMTP_ENABLED:-false}
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - SMTP_FROM=${SMTP_FROM}
    command: ["./openbooks", "server", "--port", "8080", "--persist", "--no-browser-downloads"]
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  data:
```

## Legacy Docker Compose

```docker
version: '3.3'
services:
    openbooks:
        ports:
            - '8080:80'
        volumes:
            - 'booksVolume:/books'
        restart: unless-stopped
        container_name: OpenBooks
        command: --name my_irc_name --persist
        environment:
          - BASE_PATH=/openbooks/
        image: evanbuss/openbooks:latest

volumes:
    booksVolume:
```

## Troubleshooting

### SMTP Issues
Check application logs:
```bash
docker-compose logs openbooks | grep -i smtp
```

Verify environment variables:
```bash
docker-compose exec openbooks env | grep SMTP
```

### Common Problems
- **Email not working**: Ensure `SMTP_ENABLED=true` and use App Passwords
- **Books not downloading**: Check IRC connection status in the app
- **Port conflicts**: Change host port in docker-compose.yml
- **Permission errors**: Ensure data directory is writable

### Health Check
```bash
docker-compose ps  # Should show "healthy" status
```
