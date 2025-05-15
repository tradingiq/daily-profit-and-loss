# Daily Profit and Loss Tracker for BitUnix

A lightweight system tray application that tracks and displays your daily realized profit and loss from BitUnix cryptocurrency trading.

> **Note:** This tool is currently in beta and only tested on Windows. Use at your own risk.

## Features

- Real-time tracking of daily realized profit and loss
- System tray integration with P&L display
- Desktop notifications for important events
- Automatic daily reset at midnight (Europe/Berlin timezone)
- Persistent storage of daily P&L data
- Configuration UI for API credentials and settings

## Installation

### Prerequisites

- Go 1.24 or higher
- Git

### Download

```bash
git clone https://github.com/yourusername/daily-profit-and-loss.git
cd daily-profit-and-loss
```

### Build

```bash
# Build the application
go build -o daily-pnl ./cmd/daily-pnl

# Or use make if available
make build
```

### Run

```bash
# Run directly after building
./daily-pnl

# Or install and run
go install ./cmd/daily-pnl
daily-pnl
```

## Configuration

On first run, the application will create a configuration file and prompt you to enter your BitUnix API credentials:

1. Right-click the system tray icon and select "Configure"
2. Enter your BitUnix API Key and Secret Key
3. (Optional) Change the file path for storing P&L data
4. Click "Save"

### Configuration File Location

The configuration file is stored at:
- Windows: `%APPDATA%\daily-pnl\config.json`

## Usage

Once configured, the application will:

1. Display your current daily P&L in the system tray
2. Update in real-time as your positions change
3. Reset automatically at midnight

### System Tray Options

Right-click the system tray icon to access:
- **Configure**: Update API credentials and settings
- **Info**: View version and file locations
- **Logs**: View application logs
- **Exit**: Close the application

## Troubleshooting

If you encounter issues:

1. Check the logs via the system tray menu
2. Ensure your API credentials are correct
3. Verify your internet connection
4. Restart the application

## Security Note

Your API credentials are stored locally on your machine. The application only needs read access to your BitUnix account and does not perform any trading operations.

## License

MIT License with Attribution

Copyright (c) 2025 Victor Geyer

This software is associated with Trading IQ. See the [LICENSE](LICENSE) file for details.