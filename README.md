# qrtotp

A small Go CLI tool that scans a QR code image containing an `otpauth://` URL, extracts the TOTP secret, and displays the current TOTP code.

It supports both interactive mode (with live clipboard copy and countdown) and one-shot mode (for piping into scripts).

---

## 🚀 Features

- 📸 Scan QR codes from images (PNG, JPEG) and extract secrets and provider info
- ⏱️ Show current TOTP and auto-refresh for 3 intervals (usually 90 sec)
- 📋 Automatically copies to clipboard
- 🤖 One-shot mode when used in pipelines or scripts
- 

## 🛑 Anti-Features

- 🛑 No background daemon or service
- 🛑 No config files or persistent state
- ☁️❌ Never syncs secrets to the cloud
- 📵 Runs fully offline — zero network access
- 🔒 Secrets stay local; nothing leaves your machine
- 🛰️ Can't phone home — and wouldn't even try

----
## ⚠️ Security Warning

QR code images used with this tool contain **unencrypted TOTP secrets**.

If someone gains access to those images, they can generate valid 2FA codes — just like you.

🔐 See [SECURITY.md](SECURITY.md) for for best practices for handling these files.

---
## 📦 Installation

### 🧱 Portable Go Binary

This project is written in Go and compiled as a **single statically-linked binary** named `auth`
No runtime, no dependencies — just drop it in your `$PATH` and go.

```bash
just build
cp ./auth ~/bin
```

---

## 🔧 Usage

### Try it out with the included test QR code:

```bash
just build
just run testdata/test.png

Provider: TestIssuer (test@example.com)
Current TOTP code: 492039 | Expires in: 27 sec
```

### Or from a script:

```bash
code=$(./auth testdata/test.png)
echo "TOTP: $code"

TOTP: 361121
```
