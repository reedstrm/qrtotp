# Security Policy

## 🔐 TOTP Secret Safety

This project, `qrtotp`, is a CLI tool that reads `otpauth://` QR codes containing TOTP secrets. These QR codes contain **unencrypted secrets** and must be handled as sensitive credentials.

If someone obtains a QR code image used by this tool, they can generate valid 2FA codes for the associated account.

---

## 📋 Recommendations

### ✅ Treat QR images like secrets
- Do not commit them to Git
- Do not store them unencrypted

### ✅ Store securely
  - Use encrypted folders or completely encrypted disk
  - Or keep them on **removable media** that’s stored securely and disconnected when not needed, ideally encrypted as well.


### ✅ Avoid cloud syncing
- Do not upload to Google Drive, Dropbox, OneDrive, etc. unless encrypted beforehand

### ✅ Use trusted devices
- Avoid using this tool on shared computers or remote environments

### 🔄 Rotate secrets
If you suspect that a QR code (or its extracted secret) has been exposed, rotate the 2FA key in your provider immediately.

---

## 🔒 Disclosure

This tool has no telemetry, cloud connection, or remote reporting.

If you believe you’ve found a security issue, please contact the maintainer privately.
