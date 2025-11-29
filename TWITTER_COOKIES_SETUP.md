# 🍪 Setup Twitter Scraper dengan Browser Cookies

**CARA TERBARU DAN PALING MUDAH!** Gunakan cookies dari browser Anda yang sudah login.

## ✅ Keuntungan Metode Ini

- ✅ **Paling simple** - cukup export cookies sekali
- ✅ **No Cloudflare blocking** - karena pakai cookies dari browser asli
- ✅ **No login automation** - tidak perlu worry tentang bot detection
- ✅ **Stable dan reliable** - selama cookies valid, akan terus jalan

## 🚀 Setup dalam 3 Langkah

### Step 1: Install Browser Extension

Pilih salah satu extension untuk export cookies:

**Chrome/Edge:**
- [Cookie-Editor](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
- [EditThisCookie](https://chrome.google.com/webstore/detail/editthiscookie/fngmhnnpilhplaeedifhccceomclgfbg)

**Firefox:**
- [Cookie-Editor](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)

### Step 2: Export Cookies dari Twitter/X

1. **Login ke Twitter/X** di browser Anda (https://x.com atau https://twitter.com)
2. **Pastikan sudah login** dan bisa akses Twitter normal
3. **Buka extension** Cookie-Editor (klik icon di toolbar)
4. **Export cookies**:
   - Click "Export" button  
   - Pilih format "JSON"
   - Copy atau save as file

5. **Save sebagai `cookie.json`** di root folder project:
   ```bash
   # Path yang benar:
   /home/sena/project/go-social-dls/cookie.json
   ```

**PENTING:** File harus bernama persis `cookie.json` dan berada di root folder project (sama level dengan main.go)

### Step 3: Test & Run

```bash
# Test Python script dulu
python3 scripts/twitter_scraper_cookies.py 1234567890

# Jika berhasil, build dan run server
go build -o twitter-down
./twitter-down
```

## 🧪 Testing

### Test dengan curl:

```bash
curl "http://localhost:3005/twitter?url=https://twitter.com/username/status/1234567890"
```

### Success Response:
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

## 🔄 Update Cookies (Jika Expired)

Cookies Twitter biasanya valid selama **beberapa minggu**. Jika expired:

1. **Pastikan masih login** di browser
2. **Export ulang cookies** dari extension
3. **Replace file `cookie.json`** dengan yang baru
4. **Restart server** (tidak perlu rebuild)

**Tanda cookies expired:**
- Error: "HTTP 401 Unauthorized"
- Error: "Tweet not found" untuk tweet yang sebenarnya ada
- Error: "Failed to fetch tweet"

## ⚠️ Troubleshooting

### Error: "Could not load cookie.json file"

**Penyebab:** File cookie.json tidak ditemukan atau format salah.

**Solusi:**
1. Pastikan file `cookie.json` ada di root folder project
2. Check format JSON valid dengan: `python3 -m json.tool cookie.json`
3. Re-export dari browser extension

### Error: "Missing required cookies"

**Penyebab:** Cookies tidak lengkap, missing `auth_token` atau `ct0`.

**Solusi:**
1. Pastikan Anda sudah **login** ke Twitter di browser sebelum export
2. Export cookies **dari halaman Twitter** (bukan halaman lain)
3. Re-export dengan Cookie-Editor extension

### Error: "HTTP 401 Unauthorized"

**Penyebab:** Cookies sudah expired atau tidak valid.

**Solusi:**
1. **Logout dan login ulang** di Twitter browser
2. **Export cookies baru**
3. **Replace cookie.json**
4. **Restart server**

### Error: "Tweet not found" tapi tweet ada

**Penyebab:** Cookies expired atau tweet private/deleted.

**Solusi:**
1. Check apakah tweet masih ada dan public
2. Export cookies baru dari browser
3. Test dengan tweet yang pasti public

## 🔒 Keamanan

### ⚠️ PENTING: Jangan Commit cookies.json!

File `cookie.json` berisi **credentials akun Twitter Anda**. Jangan commit ke Git!

```bash
# Add ke .gitignore
echo "cookie.json" >> .gitignore
```

### Best Practices:

1. **Gunakan akun Twitter khusus** untuk scraping (optional, tapi recommended)
2. **Jangan share file cookie.json** dengan siapapun
3. **Rotate cookies secara berkala** untuk keamanan
4. **Monitor akun** untuk aktivitas mencurigakan

## 📝 Cookies yang Dibutuhkan

Script memerlukan minimal cookies ini:
- `auth_token` - Token autentikasi utama (WAJIB)
- `ct0` - CSRF token (WAJIB)
- `guest_id` - Guest identifier (Recommended)
- `kdt` - Key derivation token (Recommended)
- Other cookies - Semakin lengkap semakin baik

Cookie-Editor extension akan export **semua cookies** secara otomatis, jadi Anda tidak perlu worry tentang mana yang perlu dan tidak.

## 🆘 Need Help?

Jika masih ada masalah:

1. ✅ Check cookies belum expired (coba browse Twitter di browser)
2. ✅ Check format JSON valid
3. ✅ Check file path benar (`cookie.json` di root project)
4. ✅ Test Python script secara manual
5. ✅ Check logs server untuk error detail

## 🔄 Cara Kerja

```
Browser (logged in) 
    ↓
Export cookies via extension
    ↓
Save as cookie.json
    ↓
Python script read cookies
    ↓
Hit Twitter GraphQL API with cookies
    ↓
Parse response & extract photo URLs
    ↓
Return to Go server
    ↓
API response to client
```

---

**Last Updated:** 2025-11-29

**Metode ini jauh lebih simple dan reliable daripada login automation!** 🎉
