# SMTP Configuration for OpenBooks

This guide explains how to set up SMTP email functionality for the "Send to Kindle" feature in OpenBooks.

## Quick Setup

1. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit the `.env` file with your SMTP settings:**
   ```bash
   nano .env
   ```

3. **Configure your SMTP provider** (see examples below)

4. **Enable SMTP:**
   ```
   SMTP_ENABLED=true
   ```

## SMTP Provider Examples

### Gmail
```env
SMTP_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com
```

**Note:** For Gmail, you must use an [App Password](https://support.google.com/accounts/answer/185833) instead of your regular password.

### Outlook/Hotmail
```env
SMTP_ENABLED=true
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587
SMTP_USERNAME=your_email@outlook.com
SMTP_PASSWORD=your_password
SMTP_FROM=your_email@outlook.com
```

### Yahoo Mail
```env
SMTP_ENABLED=true
SMTP_HOST=smtp.mail.yahoo.com
SMTP_PORT=587
SMTP_USERNAME=your_email@yahoo.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@yahoo.com
```

### Custom SMTP Server
```env
SMTP_ENABLED=true
SMTP_HOST=mail.yourdomain.com
SMTP_PORT=587
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password
SMTP_FROM=noreply@yourdomain.com
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SMTP_ENABLED` | Enable/disable SMTP functionality | `false` | Yes |
| `SMTP_HOST` | SMTP server hostname | - | Yes |
| `SMTP_PORT` | SMTP server port | `587` | No |
| `SMTP_USERNAME` | SMTP authentication username | - | Yes |
| `SMTP_PASSWORD` | SMTP authentication password | - | Yes |
| `SMTP_FROM` | Email address to send from | - | Yes |

## Security Notes

1. **Never commit your `.env` file** to version control
2. **Use App Passwords** when available (Gmail, Yahoo)
3. **Store passwords securely** in production environments
4. **Use TLS/SSL** for secure email transmission

## Testing

To test your SMTP configuration:

1. Set up your `.env` file with valid SMTP settings
2. Set `SMTP_ENABLED=true`
3. Start OpenBooks: `./openbooks server`
4. Search for a book and use the "Send to Kindle" button
5. Check the server logs for any SMTP errors

## Troubleshooting

### Common Issues

1. **Authentication failed**: Check username/password
2. **Connection timeout**: Verify host/port settings
3. **TLS errors**: Ensure port 587 for STARTTLS or 465 for SSL/TLS
4. **App Password required**: Gmail and Yahoo require app-specific passwords

### Debug Logs

OpenBooks will log SMTP attempts and errors. Check the server output for details:
```
SERVER: Send to Kindle request: Book Title by Author to user@kindle.com
```

## Kindle Setup

To receive books on your Kindle:

1. **Add the sender email** to your Amazon account's approved sender list
2. **Find your Kindle email** in your Amazon account settings
3. **Use your Kindle email** when clicking "Send to Kindle"

Your Kindle email typically looks like: `username@kindle.com`
