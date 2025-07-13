# OpenBooks - Send to Kindle Edition

A customized version of [OpenBooks](https://github.com/evan-buss/openbooks) with enhanced mobile-friendly UI and "Send to Kindle" email functionality.

## Features

### 🆕 New Features
- **Send to Kindle**: Email books directly to your Kindle device via SMTP
- **Mobile-First Design**: Card-based layout optimized for mobile devices  
- **SMTP Configuration**: Easy email setup with support for Gmail, Outlook, Yahoo, and custom SMTP servers
- **Test Email Functionality**: Built-in SMTP testing to verify your email configuration

### 📱 UI Improvements
- Modern card-based book display instead of table rows
- Responsive design that works great on mobile devices
- Simplified interface without sidebar/history clutter
- Clean notification system

## Quick Start

1. **Download a release** or build from source
2. **Configure SMTP** (see [SMTP Setup](#smtp-setup))
3. **Run the application**:
   ```bash
   ./openbooks server --name your_username --browser
   ```
4. **Search for books** and use the "Send to Kindle" button!

## SMTP Setup

Create a `.env` file in the same directory as your OpenBooks binary:

```env
SMTP_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com
```

For detailed SMTP configuration instructions, see [SMTP_README.md](SMTP_README.md).

## Building from Source

### Prerequisites
- Go 1.19+
- Node.js 16+
- npm or yarn

### Build Steps

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-username/openbooks-kindle.git
   cd openbooks-kindle
   ```

2. **Build the frontend**:
   ```bash
   cd server/app
   npm install
   npm run build
   cd ../..
   ```

3. **Build the backend**:
   ```bash
   go build -o openbooks ./cmd/openbooks/
   ```

4. **Run the application**:
   ```bash
   ./openbooks server --name your_username --browser
   ```

## Usage

1. **Start the server** with your IRC username
2. **Open your browser** to http://localhost:5228
3. **Search for books** using the search bar
4. **Send to Kindle** by clicking the email button on any book card
5. **Test SMTP** using the envelope icon in the top header

## Configuration

### Command Line Options
```bash
./openbooks server --help
```

### Environment Variables
See [SMTP_README.md](SMTP_README.md) for complete SMTP configuration options.

## Contributing

This is a fork of the original OpenBooks project with custom modifications. For the original project, visit [evan-buss/openbooks](https://github.com/evan-buss/openbooks).

### Changes Made
- Converted table-based UI to mobile-friendly cards
- Added SMTP email functionality  
- Removed sidebar and search history
- Added test email functionality
- Enhanced responsive design

## License

This project maintains the same license as the original OpenBooks project.

## Credits

- Original OpenBooks project: [evan-buss/openbooks](https://github.com/evan-buss/openbooks)
- UI framework: [Mantine](https://mantine.dev/)
- Icons: [Phosphor Icons](https://phosphoricons.com/)

## Support

For issues related to the Send to Kindle functionality or mobile UI, please open an issue in this repository.

For general OpenBooks issues, refer to the [original project](https://github.com/evan-buss/openbooks).
   - Linux users may have to run `chmod +x [binary name]` to make it executable
3. `./openbooks --help`
   - This will display all possible configuration values and introduce the two modes; CLI or Server.

### Docker

- Basic config
  - `docker run -p 8080:80 evanbuss/openbooks`
- Config to persist all eBook files to disk
  - `docker run -p 8080:80 -v /home/evan/Downloads/openbooks:/books evanbuss/openbooks --persist`

### Setting the Base Path

OpenBooks server doesn't have to be hosted at the root of your webserver. The basepath value allows you to host it behind a reverse proxy. The base path value must have opening and closing forward slashes (default "/").

- Docker
  - `docker run -p 8080:80 -e BASE_PATH=/openbooks/ evanbuss/openbooks`
- Binary
  - `./openbooks server --basepath /openbooks/`

## Usage

For a complete list of features use the `--help` flags on all subcommands.
For example `openbooks cli --help or openbooks cli download --help`. There are
two modes; Server or CLI. In CLI mode you interact and download books through
a terminal interface. In server mode the application runs as a web application
that you can visit in your browser.

Double clicking the executable will open the UI in your browser. In the future it may use [webviews](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to provide a "native-like" desktop application. 

## Development

### Install the dependencies

- `go get`
- `cd server/app && npm install`
- `cd ../..`
- `go run main.go`

### Build the React SPA and compile binaries for multiple platforms.

- Run `./build.sh`
- This will install npm packages, build the React app, and compile the executable.

### Build the go binary (if you haven't changed the frontend)

- `go build`

### Mock Development Server

- The mock server allows you to debug responses and requests to simplified IRC / DCC
  servers that mimic the responses received from IRC Highway.
- ```bash
  cd cmd/mock_server
  go run .
  # Another Terminal
  cd cmd/openbooks
  go run . server --server localhost --log
  ```

### Desktop App
Compile OpenBooks with experimental webview support:

``` shell
cd cmd/openbooks
go build -tags webview
```


## Why / How

- I wrote this as an easier way to search and download books from irchighway.net. It handles all the extraction and data processing for you. You just have to click the book you want. Hopefully you find it much easier than the IRC interface.
- It was also interesting to learn how the [IRC](https://en.wikipedia.org/wiki/Internet_Relay_Chat) and [DCC](https://en.wikipedia.org/wiki/Direct_Client-to-Client) protocols work and write custom implementations.

## Technology

- Backend
  - Golang
  - Chi
  - gorilla/websocket
  - Archiver (extract files from various archive formats)
- Frontend
  - React.js
  - TypeScript
  - Redux / Redux Toolkit
  - Mantine UI / @emotion/react
  - Framer Motion
