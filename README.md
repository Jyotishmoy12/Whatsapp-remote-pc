# 🛠️ WhatsApp Remote PC Gateway: Setup Guide

This guide covers how to set up a secure, Go-based remote administration tool that allows you to control your Windows PC via WhatsApp.

## 1. Prerequisites & Environment Setup

To keep the bot secure, we use an `.env` file to store sensitive credentials.

**Create an .env file** in your project root with the following variables:

```
AUTHORIZED_NUMBER: Your full phone number with country code (e.g., 919876543210@s.whatsapp.net).
OWNER_LID: Your WhatsApp LID (found in the terminal logs during the first run).
```

**Install Dependencies:**

- **whatsmeow**: For the WhatsApp protocol.
- **kbinani/screenshot**: For capturing the desktop.
- **joho/godotenv**: For loading your .env variables.
-
## 2. The Commands

Once the bot is running, you can send these commands to yourself on WhatsApp:

| Command | Action |
|---------|--------|
| `!status` | Checks if the PC is online and the bot is active. |
| `!ls [path]` | Lists all files/folders in the current or specified directory. |
| `!cd [path]` | Changes the working directory for future commands. |
| `!find [file]` | Recursively searches for a file starting from the current directory. |
| `!screen` | Sends a high-resolution screenshot of your current Windows desktop. |
| `!get [path]` | Downloads the specified file from your PC to your phone. |
| `!cmd [command]` | Executes any terminal command (e.g., go build, ipconfig, git status). |
| `!restart` | Restarts the Windows PC. |
| `!shutdown` | Shuts down the Windows PC. |
| `!reset` | The Nuclear Option: Wipes the session, deletes local data, and kills the bot. |## 3. Background Automation (The "Always-On" Mode)

To ensure the bot is always ready without keeping a terminal open, follow these steps:

### A. Compile the "Stealth" Binary

Run this command in your project folder to create an invisible background process:

```bash
go build -ldflags "-H windowsgui" -o WhatsAppRemote.exe
```

### B. Configure Windows Task Scheduler

1. **Create Task**: Name it WhatsAppRemotePC and check "Run with highest privileges".
2. **Trigger**: Set to "At log on" for your user account.
3. **Action**:
   - Program/script: Path to your WhatsAppRemote.exe.
   - Start in: The exact folder path containing your .env and session.db (No quotes, no trailing slash).
4. **Power**: In Windows Power Options, set "When I close the lid" to "Do nothing" so the bot stays active when the laptop is closed.## 4. Important Security Notes

- **Authorized Access**: The bot will only respond to the `AUTHORIZED_NUMBER` and `OWNER_LID` defined in your `.env`. All other messages are ignored.
- **Session Management**: The `session.db` file holds your login token. If you ever suspect a breach, use the `!reset` command immediately to revoke access.
- **Firewall**: Ensure your firewall allows Go binaries to access the network so the WhatsApp WebSocket can maintain a connection.
