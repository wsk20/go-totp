# Multi-Account TOTP Manager

A **multi-account TOTP (Time-based One-Time Password) manager** built with Go, supporting:

- Adding accounts via `otpauth://` URI  
- Manually adding accounts (username + secret key)  
- Deleting and listing accounts  
- Verifying input codes  
- Dynamically displaying multiple TOTP codes with countdowns  
- Supporting various algorithms (SHA1/SHA256/SHA512) with configurable period and digits  
- Cross-platform, storing account information locally at `~/.totp_accounts.json`  

---

## Installation

Make sure Go 1.18+ is installed, then use `go install`:

```bash
go install github.com/wsk20/go-totp@latest
````

After installation, you can run the CLI directly:

```bash
go-totp
```

---

## Usage Examples

### 1. Add an account (URI)

```bash
go-totp --add "otpauth://totp/label?secret=ABC123&issuer=Example&algorithm=SHA1&period=30&digits=6"
```

Example output:

```
✅ Successfully added: label
```

### 2. Add an account (manual)

```bash
go-totp --add-user alice --add-key ABC123 --add-issuer Example --add-algo SHA1 --add-period 30 --add-digits 6
```

### 3. Remove an account

```bash
go-totp --remove alice
```

### 4. List all accounts

```bash
go-totp --list
```

Example output:

```
- alice (Example) [SHA1]
- bob (Google) [SHA1]
```

### 5. Show or verify a specific account

```bash
go-totp --account alice
```

```bash
go-totp --account alice --verify 123456
```

### 6. Run dynamic TOTP display

```bash
go-totp
```

* Supports displaying multiple accounts simultaneously
* Real-time countdown, with a `beep` alert near expiration
* Supports Ctrl+C to exit

---

## Dynamic Display Example

When running, the terminal displays:

```
🔐 Multi-Account Dynamic TOTP Manager
========================================
Issuer: Example
Account: alice
Algorithm: SHA1 | Period: 30s
Code: 123456
Remaining time: 25s [████████░░░░░░░░░░]

Issuer: Google
Account: bob
Algorithm: SHA1 | Period: 30s
Code: 654321
Remaining time: 12s [█████████████░░░░░]
```

* Codes and remaining time update every second
* Beeps when remaining time ≤ 5s
* Automatically arranges any number of accounts

---

## Command-Line Flags

| Flag           | Description                                       |
| -------------- | ------------------------------------------------- |
| `--add`        | Add account via URI (`otpauth://totp/...`)        |
| `--remove`     | Remove account by label                           |
| `--list`       | List all accounts                                 |
| `--verify`     | Verify an input code                              |
| `--account`    | Specify account(s), comma-separated               |
| `--add-user`   | Add account username (manual)                     |
| `--add-key`    | Add account secret (manual)                       |
| `--add-issuer` | Issuer / platform name                            |
| `--add-algo`   | Hash algorithm: SHA1/SHA256/SHA512 (default SHA1) |
| `--add-period` | Time step in seconds (default 30)                 |
| `--add-digits` | Code digits (default 6)                           |

---

## File Storage

Accounts are stored in the user’s home directory:

```
~/.totp_accounts.json
```

* Automatically deduplicated
* JSON format, easy to backup or migrate

---

## ANSI Color Display

* ✅ Success: Green
* ⚠️ Warning: Yellow
* ❌ Error: Red
* Dynamic countdown displayed with colored progress bars

---

## Notes

* Only supports TOTP (does not support HOTP)
* Ctrl+C restores cursor and clears the screen
* Supports SHA1/SHA256/SHA512 algorithms

---

## License

MIT License