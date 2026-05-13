<div align="center">

<img src="https://authvaultix.com/assets/img/logo.webp" alt="AuthVaultix Logo" width="80" height="80" />

# AuthVaultix Go Example

**A complete, ready-to-use Go CLI integration example for the [AuthVaultix](https://authvaultix.com) authentication platform.**

[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![AuthVaultix](https://img.shields.io/badge/AuthVaultix-API%201.2-6366F1?style=for-the-badge)](https://authvaultix.com)
[![License](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge)](LICENSE)
[![Discord](https://img.shields.io/badge/Discord-Join%20Us-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/muHy3qxcub)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white)](https://www.microsoft.com/windows)

</div>

---

## 📖 Overview

This repository provides a **plug-and-play Go CLI example** demonstrating how to integrate the [AuthVaultix](https://authvaultix.com) authentication API into your Go application. It includes:

- 🔐 **Login** — Authenticate users with username & password
- 📝 **Register** — Create new accounts with a license key
- 🔑 **License Login** — Access the app using only a license key
- 🖥️ **HWID Detection** — Automatically reads the Windows User SID via PowerShell
- 📊 **User Info Display** — Shows username, IP, HWID & active subscriptions with expiry
- 🌐 **HTTPS API v1.0** — Communicates with `https://authvaultix.com/api/1.0/`

> Built as a **terminal-based interactive CLI app** using only Go's standard library (no external HTTP dependencies needed beyond `encoding/json` and `net/http`).

---

## 🗂️ Project Structure

```
authvaultix-go-example/
├── main.go             # Entry point — interactive CLI menu
├── authvaultix.go      # AuthVaultix API wrapper (core library)
└── run.bat             # Quick setup & run script for Windows
```

---

## ⚡ Quick Start

### 1. Prerequisites

- [Go](https://golang.org/dl/) **1.21 or higher**
- Windows OS (HWID detection uses PowerShell)
- An **AuthVaultix** account → [Register here](https://authvaultix.com)

### 2. Clone the Repository

```bash
git clone https://github.com/AuthVaultix/AuthVaultix-Go-Example.git
cd AuthVaultix-Go-Example
```

### 3. Configure Your Credentials

Open `main.go` and fill in your application details from the [AuthVaultix Dashboard](https://authvaultix.com):

```go
AuthVaultixApp := NewAuthVaultix(
    "YourAppName",   // name
    "your_ownerid",  // ownerid
    "your_secret",   // secret
    "1.0",           // version
)
```

> ⚠️ **Never commit real credentials to a public repository.**

### 4. Initialize Go Module & Run

```bash
go mod init authvaultix
go run .
```

Or simply use the included Windows batch script:

```batch
run.bat
```

---

## 🖥️ CLI Usage

When you run the app, you'll see an interactive menu:

```
Connecting...
✅ Initialized Successfully!

[1] Login
[2] Register
[3] License Login
[4] Exit
Choose option:
```

#### Login (Option 1)

```
Username: johndoe
Password: ••••••••

✅ Logged in!

👤 User Info:
  Username: johndoe
  IP: 192.168.x.x
  HWID: S-1-5-21-...
  → default | Expiry: 1767225600 | Left: 20736000
```

#### Register (Option 2)

```
Username: newuser
Password: ••••••••
License: XXXX-XXXX-XXXX-XXXX

✅ Registered Successfully!
```

#### License Login (Option 3)

```
License: XXXX-XXXX-XXXX-XXXX

✅ License Login Successful!
```

---

## 🧩 Library Usage (`authvaultix.go`)

You can use the `Authvaultix` struct directly in your own Go project:

### Initialize

```go
app := NewAuthVaultix("AppName", "ownerid", "secret", "1.0")
app.Init() // Must be called before any other method
```

### Login

```go
app.Login("username", "password")
```

### Register

```go
app.Register("username", "password", "LICENSE-KEY", "")
```

### License Login

```go
app.LicenseLogin("LICENSE-KEY")
```

---

## ⚙️ API Reference

### Authentication & Session
| Method | Returns | Description |
|---|---|---|
| `Init()` | `bool` | Initializes the session with the API. |
| `Login(username, pass)` | `bool` | Authenticates a user. |
| `Register(user, pass, key, email)` | `bool` | Registers a new user. |
| `LicenseLogin(licenseKey)` | `bool` | Authenticates directly via license key. |
| `Check()` | `bool` | Validates the current session. |
| `Logout()` | | Terminates session. |

### Account Management
| Method | Returns | Description |
|---|---|---|
| `Upgrade(username, licenseKey)` | `bool` | Upgrades user's subscription. |
| `ForgotPassword(username, email)`| `bool` | Triggers a password reset email. |
| `ChangeUsername(newUsername)` | | Changes the current user's username. |

### Security & Logging
| Method | Returns | Description |
|---|---|---|
| `Ban(reason)` | `bool` | Bans the currently authenticated user. |
| `CheckBlacklist()` | `bool` | Checks if the current HWID is blacklisted. |
| `Log(message)` | `bool` | Sends a log message to the dashboard. |

### Variables & Data
| Method | Returns | Description |
|---|---|---|
| `GetGlobalVar(varId)` | `string` | Fetches a global server variable. |
| `GetVar(varName)` | `string` | Fetches a user-specific variable. |
| `SetVar(varName, value)` | `bool` | Sets a user-specific variable. |
| `Download(fileId)` | `[]byte` | Securely downloads a file. |

### Communication
| Method | Returns | Description |
|---|---|---|
| `FetchOnline()` | `[]OnlineUser` | Retrieves a list of online clients. |
| `ChatSend(message, channel)` | `bool` | Sends a chat message. |
| `ChatFetch(channel)` | `[]ChatMessage`| Fetches chat history for a channel. |

---

## 📦 Dependencies

This project uses **only the Go standard library** — no `go get` required for the core SDK.

| Package                         | Source | Purpose                       |
| ------------------------------- | ------ | ----------------------------- |
| `net/http`                      | stdlib | HTTP POST requests to the API |
| `encoding/json`                 | stdlib | JSON parsing of API responses |
| `os/exec`                       | stdlib | PowerShell HWID detection     |
| `bytes`, `strings`, `fmt`, `io` | stdlib | Utilities                     |

> The `run.bat` calls `go mod init authvaultix` automatically — no manual module setup needed.

---

## 🖥️ HWID Detection

This example uses your **Windows User SID** (Security Identifier) as the hardware fingerprint, fetched via PowerShell:

```powershell
[System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
```

If the command fails for any reason, it falls back to `"UNKNOWN_HWID"`.

> For cross-platform support, you can replace this with a library like [`github.com/denisbrodbeck/machineid`](https://github.com/denisbrodbeck/machineid).

---

## 🔒 Security Notes

| Concern                  | Recommendation                                                           |
| ------------------------ | ------------------------------------------------------------------------ |
| Credentials in `main.go` | Use environment variables (`os.Getenv`) or a config file in production   |
| HWID Binding             | AuthVaultix locks sessions to the detected SID by default                |
| HTTPS                    | All API calls go to `https://authvaultix.com` — always encrypted         |
| URL Encoding             | Use `url.Values` instead of manual string building for production safety |

---

## 🛠️ Customization

- **Add 2FA support**: Extend the `Login()` and `License()` payload maps with a `"code"` key
- **Subscription gating**: Loop through `user.Subscriptions` and check for a specific `sub.Subscription` name before granting app access
- **Cross-platform HWID**: Replace PowerShell logic with [`machineid`](https://github.com/denisbrodbeck/machineid) for Linux/macOS support
- **GUI**: Build a desktop frontend using [`Fyne`](https://fyne.io/) or [`Wails`](https://wails.io/) on top of this core library

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the repository
2. Create a new branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -m 'Add my feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Open a Pull Request

---

## 💬 Support

- 📖 [AuthVaultix Documentation](https://authvaultix.com)
- 💬 [Discord Community](https://discord.gg/muHy3qxcub)
- 🐛 [Open an Issue](https://github.com/YOUR_USERNAME/authvaultix-go-example/issues)

---

## 📄 License

This project is licensed under the **MIT License** — feel free to use, modify, and distribute it.

---

<div align="center">

Made with 🐹 Go + ❤️ using [AuthVaultix](https://authvaultix.com)

</div>
