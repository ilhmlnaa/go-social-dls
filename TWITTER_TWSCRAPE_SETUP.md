# 🐦 Setup Twitter Scraper dengan twscrape

Guide lengkap setup dan konfigurasi twscrape untuk endpoint Twitter.

## 📋 Prerequisites

- Python 3.8 atau lebih tinggi
- pip (Python package manager)
- Akun Twitter (untuk scraping)

## 🚀 Instalasi

### 1. Install Python Dependencies

```bash
# Install twscrape
pip3 install twscrape

# Atau jika ada permission error, gunakan --user
pip3 install --user twscrape
```

### 2. Verifikasi Instalasi

```bash
python3 -c "import twscrape; print('twscrape version:', twscrape.__version__)"
```

Jika berhasil, akan menampilkan versi twscrape yang terinstall.

## 🔐 Konfigurasi Akun Twitter

### Cara 1: Menggunakan Script Setup (Recommended)

```bash
# Tambah akun Twitter
python3 scripts/setup_twscrape.py add <username> <password> <email> <email_password>

# Contoh:
python3 scripts/setup_twscrape.py add mytwitter mypass123 [email protected] emailpass123
```

**Parameter:**
- `username`: Username Twitter Anda (tanpa @)
- `password`: Password Twitter
- `email`: Email yang terdaftar di Twitter
- `email_password`: Password email (untuk verifikasi 2FA jika diperlukan)

### Cara 2: Menggunakan twscrape CLI

```bash
# Tambah akun
twscrape add_accounts accounts.txt username:password:email:email_password

# Login semua akun
twscrape login_accounts
```

## ✅ Verifikasi Setup

### Cek Akun yang Terkonfigurasi

```bash
python3 scripts/setup_twscrape.py list
```

Output yang diharapkan:
```
📊 Total accounts: 1

Username: mytwitter
Status: ACTIVE
Locks: {}
--------------------------------------------------
```

### Test Manual Script

```bash
# Test dengan tweet ID
python3 scripts/twitter_scraper.py 1234567890

# Jika berhasil, akan return JSON seperti:
# {
#   "success": true,
#   "tweet_id": "1234567890",
#   "photos": ["https://pbs.twimg.com/media/xxx.jpg"],
#   "photo_count": 1
# }
```

## 🔄 Restart Server Go

Setelah setup Python selesai, rebuild dan restart server Go:

```bash
# Build Go server
go build -o twitter-down

# Run server
./twitter-down
```

## 🧪 Test Endpoint

Test endpoint Twitter dengan curl:

```bash
curl "http://localhost:3005/twitter?url=https://twitter.com/username/status/1234567890"
```

Response yang diharapkan:
```json
{
  "success": true,
  "message": "Berhasil mengambil 2 gambar",
  "data": [
    "https://pbs.twimg.com/media/xxx.jpg?name=large",
    "https://pbs.twimg.com/media/yyy.jpg?name=large"
  ]
}
```

## ⚠️ Troubleshooting

### Error: "No Twitter accounts configured"

**Penyebab:** Belum ada akun Twitter yang ditambahkan.

**Solusi:**
```bash
python3 scripts/setup_twscrape.py add <username> <password> <email> <email_password>
```

### Error: "Login failed" atau "Bad authentication data"

**Penyebab:** Username, password, atau email salah.

**Solusi:**
1. Cek kembali credentials Anda
2. Pastikan akun tidak ter-lock atau suspended
3. Jika ada 2FA, pastikan email_password benar

### Error: "Rate limit exceeded"

**Penyebab:** Terlalu banyak request dalam waktu singkat.

**Solusi:**
1. Tunggu beberapa menit
2. Tambah lebih banyak akun Twitter untuk rotate
3. Implement rate limiting di aplikasi

### Error: "python3: command not found"

**Penyebab:** Python tidak terinstall atau tidak ada di PATH.

**Solusi:**
```bash
# Install Python di Ubuntu/Debian
sudo apt update
sudo apt install python3 python3-pip

# Install Python di macOS
brew install python3
```

### Error: "Module 'twscrape' not found"

**Penyebab:** twscrape belum terinstall.

**Solusi:**
```bash
pip3 install twscrape

# Jika masih error, coba:
python3 -m pip install twscrape
```

## 🔒 Keamanan

### Best Practices

1. **Jangan hardcode credentials** di code
2. **Gunakan akun khusus** untuk scraping (bukan akun pribadi)
3. **Rotate accounts** untuk menghindari rate limit
4. **Respect rate limits** Twitter

### Protect Credentials

Jangan commit file yang berisi credentials:

```bash
# Add ke .gitignore
echo "accounts.txt" >> .gitignore
echo "scripts/__pycache__/" >> .gitignore
```

## 📊 Multiple Accounts (Advanced)

Untuk menghindari rate limit, tambah beberapa akun:

```bash
# Tambah akun pertama
python3 scripts/setup_twscrape.py add acc1 pass1 [email protected] emailpass1

# Tambah akun kedua
python3 scripts/setup_twscrape.py add acc2 pass2 [email protected] emailpass2

# Tambah akun ketiga
python3 scripts/setup_twscrape.py add acc3 pass3 [email protected] emailpass3
```

twscrape akan otomatis rotate antar akun untuk menghindari rate limit.

## 🆘 Need Help?

Jika masih ada masalah:

1. Cek logs di console saat run server
2. Test Python script secara manual
3. Pastikan Python version >= 3.8
4. Cek dokumentasi twscrape: https://github.com/vladkens/twscrape

## 📝 Notes

- twscrape menyimpan data akun di `~/.twscrape/accounts.db`
- Setiap akun punya rate limit sendiri (~500 tweets/15 min)
- Login session bertahan beberapa hari, tidak perlu login ulang setiap request
- Akun Twitter harus aktif (tidak suspended/limited)

---

**Last Updated:** 2025-11-29
