# Configuration

GoDown uses environment variables for configuration. You can set these variables in your environment or create a `.env` file.

## Configuration Options

### Application Settings

- `APP_NAME`: Application name (default: "cheetah")
- `APP_VERSION`: Application version (default: "1.0.0")
- `APP_ENV`: Environment (development/production) (default: "development")
- `REPO_DIR`: Directory for downloaded files (default: "repo")

### Download Settings

- `DOWNLOAD_BATCH_SIZE`: Number of concurrent downloads per batch (default: 50)
- `DOWNLOAD_MAX_ERRORS`: Maximum errors before stopping (default: 100)
- `DOWNLOAD_ROUTINE_ERR_MAX`: Maximum retries per file (default: 10)

### HTTP Settings

- `HTTP_TIMEOUT`: HTTP request timeout (default: "30s")
- `HTTP_MAX_RETRIES`: Maximum HTTP retries (default: 3)
- `HTTP_RETRY_DELAY`: Delay between retries (default: "1s")
- `HTTP_MAX_CONCURRENCY`: Maximum concurrent connections (default: 50)
- `HTTP_USER_AGENT`: User agent string for HTTP requests

### Logging Settings

- `LOG_LEVEL`: Log level (debug/info/warn/error) (default: "info")
- `LOG_PRETTY`: Enable pretty printing for logs (default: true)

## Example Usage

```bash
# Set environment variables
export DOWNLOAD_BATCH_SIZE=100
export LOG_LEVEL=debug
export HTTP_TIMEOUT=60s

# Run the application
./bin/cheetah
```

Or create a `.env` file (copy from `config.example.env`):

```bash
cp config/config.example.env .env
# Edit .env file with your preferred settings
```
