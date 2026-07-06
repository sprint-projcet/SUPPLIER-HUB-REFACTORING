**Laporan Analisis dan Refactoring Kode**

**Aplikasi Web B2B Marketplace SupplierHub**

**Mata Kuliah: Rekayasa Perangkat Lunak II**

![](./image1.png){width="5.129861111111111in"
height="1.8660028433945757in"}

**Dosen Pengampu**

Muhammad Yusril Helmi Setyawan, S. Kom., M. Kom., SFPC.

**Disusun oleh :**

Zahra Ramanayshilla Sopian 714240003

Keyla Sun 714240048

Zahra Nur'azijah Lutfiani 714240052

**Kelas : 2A D4 Teknik Informatika**

**PROGRAM STUDI DIV TEKNIK INFORMATIKA**

**SCHOOL OF INFORMATION TECHNOLOGY**

**UNIVERSITAS LOGISTIK DAN BISNIS INTERNASIONAL**

**BANDUNG**

**2026\
**

# 1. Identitas Proyek

+-------------------+--------------------------------------------------+
| **Komponen**      | **Isi**                                          |
+===================+==================================================+
| **Nama Aplikasi** | SupplierHub                                      |
+-------------------+--------------------------------------------------+
| **Jenis           | Marketplace B2B (Business-to-Business)           |
| Aplikasi**        |                                                  |
+-------------------+--------------------------------------------------+
| **Pola            | RESTful API dengan arsitektur MVC / Clean        |
| Arsitektur**      | Architecture di sisi backend                     |
+-------------------+--------------------------------------------------+
| **Teknologi       | Golang (Gin Web Framework), GORM, MySQL, HTML5   |
| Utama**           | Statis, TailwindCSS CDN, Vanilla JavaScript      |
+-------------------+--------------------------------------------------+
| **Topik           | MVC, SOLID, Clean Code, High Cohesion, Low       |
| Praktikum**       | Coupling, dan Refactoring                        |
+-------------------+--------------------------------------------------+
| **Nama Kelompok** | Kelompok 4 (SupplierHub)                         |
+-------------------+--------------------------------------------------+
| **Anggota         | Keyla Sun                                        |
| Kelompok**        |                                                  |
|                   | Zahra Nur'azijah Lutfiani                        |
|                   |                                                  |
|                   | Zahra Ramanayshilla Sopian                       |
+-------------------+--------------------------------------------------+
| **Repository**    | ht                                               |
|                   | tps://github.com/sprint-projcet/SUPPLIER-HUB.git |
+-------------------+--------------------------------------------------+
| **Tanggal         | 21 Juni 2026                                     |
| Revisi**          |                                                  |
+-------------------+--------------------------------------------------+

# 2. Deskripsi Singkat Aplikasi

SupplierHub adalah platform digital berbasis web yang dirancang sebagai
ekosistem penghubung pasokan bahan baku antara UMKM (sebagai pembeli)
dengan pihak Supplier (sebagai penyedia barang/stok tangan pertama).
Platform ini menerapkan tiga algoritma inti pada umkm_controller.go:
Knuth-Morris-Pratt untuk pencarian string, Quick Sort untuk pengurutan
harga responsif, dan Binary Search untuk validasi biner instan. Platform
juga menerapkan pemotongan fee layanan sebesar 3% dari setiap transaksi
lunas sebagai model monetisasi platform.

# 3. Ruang Lingkup Analisis Kode

Analisis dilakukan secara menyeluruh terhadap berkas-berkas pengontrol
utama, entry point backend, konfigurasi variabel lingkungan, skema model
database, dan script fetch logika visualisasi data pada antarmuka
front-end portal.

Lapisan Backend:

-   backend/main.go

-   controllers/auth_controller.go

-   controllers/umkm_controller.go

-   controllers/supplier_controller.go

-   controllers/admin_controller.go

-   controllers/payment_controller.go

-   models/models.go

Lapisan Frontend:

-   Login/login.html (Logika autentikasi & Google OAuth)

-   dashboard/admin.html (Logika diagram analitik)

-   assets/js/auth.js

# 4. Ringkasan Temuan Masalah Kode (Tabel Matriks)

  -------------------------------------------------------------------------------------------------------------
  **No**   **Lokasi Kode (File &         **Masalah Utama yang Ditemukan**     **Prinsip       **Klasifikasi
           Fungsi)**                                                          Desain          Dampak**
                                                                              Terkait**       
  -------- ----------------------------- ------------------------------------ --------------- -----------------
  1        auth_controller.go (Register) God Function / Fat Controller:       SOLID (SRP) &   High
                                         pencampuran validasi, upload file    Clean Code      Maintainability
                                         disk, bcrypt hashing, dan query DB                   Risk

  2        auth_controller.go            Kunci rahasia cadangan hardcoded     Secure          Security
           (createAuthToken)             (\"super_secret_key_supplierhub\")   Configuration / Vulnerability
                                                                              12-Factor App   

  3        auth_controller.go (Login)    Validasi format email manual         Clean Code (DRY Code Duplication
                                         menggunakan kompilasi Regex lokal    / Reusability)  

  4        umkm_controller.go            Masalah performa N+1 Query akibat    Performance /   Server
           (GetProducts)                 eksekusi agregasi sub-query rating   Database        Performance
                                         di dalam perulangan                  Efficiency      Degradation

  5        umkm_controller.go            Penarikan seluruh ID produk ke RAM   Clean Code      High Resource
           (CreateOrder)                 (Memory Bloat) dan penggunaan        (Magic Value) & Consumption
                                         desimal hardcoded (0.03)             Opt. Memori     

  6        umkm_controller.go (KMPMatch) Penulisan fungsi detil matematika    Separation of   Low Cohesion
                                         algoritma KMP menyatu dalam file     Concerns (SoC)  
                                         handler                                              

  7        supplier_controller.go        Ketergantungan kaku handler langsung SOLID (DIP) /   Untestable Code
           (CreateProduct)               ke objek database global config.DB   Coupling        

  8        supplier_controller.go        Fungsi pemetaan payload respons JSON Clean Code      Code Bloat
           (supplierProfilePayload)      lokal mengotori ranah operasional    (Domain         
                                         HTTP context                         Cohesion)       

  9        supplier_controller.go        Duplikasi logika pengondisian string Clean Code      Data
           (UpdateOrderStatus)           untuk penyeragaman format status     (DRY)           Inconsistency
                                         pesanan                                              Risk

  10       admin_controller.go           Rantai statement SQL GORM kompleks   Low Coupling /  Architectural
           (GetAdminSuppliers)           (Like Search) mencemari lapisan      Layered         Pollution
                                         penyajian data                       Architecture    

  11       admin_controller.go           Deklarasi data array statis status   Clean Code      Low Code
           (GetAdminStats)               pesanan aktif yang panjang di dalam  (Readability)   Readability
                                         inti fungsi                                          

  12       admin_controller.go           Penempatan fungsi utilitas konversi  High Cohesion   Architecture
           (toString)                    tipe data generik di level                           Violations
                                         pengontrol administratif                             

  13       payment_controller.go         Instansiasi http client secara       SOLID (DIP) /   Integration
           (forwardPaymentToSmartBank)   langsung, menghalangi teknik mocking Testability     Testing Blocker
                                         network                                              

  14       main.go (main)                Nilai parameter nomor port web       Cloud Native /  Inflexible
                                         server dikunci statis pada string    Twelve-Factor   Deployment
                                         \":8080\"                            App             

  15       Login/login.html &            Duplikasi state management token     State Integrity Authentication
           assets/js/auth.js             (pencampuran localStorage.token dan  / Shared        Bug Risk
                                         user_session)                        Session Helper  
  -------------------------------------------------------------------------------------------------------------

**5. Analisis Detail Before-After Refactoring (15 Temuan Masalah)**

## Temuan 1: Fat Controller & Magic Value pada Fitur Registrasi

**[Lokasi Kode]{.underline}**

> auth_controller.go :: Register()

**[Kode Sebelum Refactoring]{.underline}**

> func Register(c \*gin.Context) {
>
> os.MkdirAll(\"uploads/documents\", os.ModePerm)
>
> filepath := \"uploads/documents/\" + filename
>
> // Proses upload file\...
>
> hashedPassword, \_ := bcrypt.GenerateFromPassword(\[\]byte(password),
> 10)
>
> config.DB.Create(&newUser)
>
> }

**[Masalah yang Ditemukan]{.underline}**

Fungsi Register bertindak sebagai God Function. Menangani pembacaan
context, manipulasi folder lokal statis dengan path hardcoded
(\"uploads/documents\"), enkripsi data bcrypt, hingga penulisan SQL
database secara langsung dalam satu fungsi.

**[Prinsip Desain yang Dilanggar]{.underline}**

SOLID (Single Responsibility Principle): setiap fungsi seharusnya hanya
memiliki satu alasan untuk berubah. Clean Code (Magic Value): nilai
string literal hardcoded di tengah kode bisnis rawan typo dan sulit
diubah.

**[Kode Sesudah Refactoring]{.underline}**

> // Di config/app.go - konstanta dipusatkan
>
> const DocumentUploadDir = \"uploads/documents\"
>
> // Di auth_controller.go - controller hanya fokus pada HTTP layer
>
> func Register(c \*gin.Context) {
>
> var input dto.RegisterInput
>
> if err := c.ShouldBind(&input); err != nil { \... }
>
> err := services.NewUserService().RegisterNewUser(input)
>
> if err != nil { c.JSON(500, gin.H{\"error\": err.Error()}); return }
>
> c.JSON(201, gin.H{\"message\": \"Registrasi UMKM/Supplier berhasil\"})
>
> }

**[Dampak Perbaikan]{.underline}**

Controller menjadi sangat ramping dan hanya bertanggung jawab pada HTTP
layer. Logika bisnis pendaftaran dapat diuji secara mandiri melalui unit
testing tanpa memerlukan HTTP context maupun database asli.

## Temuan 2: Celah Keamanan Hardcoded Fallback JWT Secret Key

**[Lokasi Kode]{.underline}**

> auth_controller.go :: createAuthToken()

**[Kode Sebelum Refactoring]{.underline}**

> secret := os.Getenv(\"JWT_SECRET\")
>
> if secret == \"\" {
>
> secret = \"super_secret_key_supplierhub\"
>
> }

**[Masalah yang Ditemukan]{.underline}**

Menyediakan string fallback secret bawaan sangat berbahaya. Jika
aplikasi di-deploy ke produksi tanpa menyetel .env, token JWT menjadi
mudah dipalsukan karena kunci rahasianya dapat ditemukan langsung di
source code.

**[Prinsip Desain yang Dilanggar]{.underline}**

Secure Configuration Management / Twelve-Factor App: konfigurasi
deployment harus dapat diubah dari lingkungan tanpa modifikasi kode
sumber. Kode tidak boleh menyediakan kunci kriptografi default.

**[Kode Sesudah Refactoring]{.underline}**

> // Di config/app.go - dieksekusi saat aplikasi pertama kali menyala
>
> if os.Getenv(\"JWT_SECRET\") == \"\" {
>
> log.Fatal(\"Kritis: Environment variable JWT_SECRET tidak
> ditemukan!\")
>
> }

**[Dampak Perbaikan]{.underline}**

Sistem menolak berjalan dengan kunci default. Pengembang dipaksa untuk
selalu mengonfigurasi JWT_SECRET dengan benar sebelum aplikasi dapat
dijalankan di lingkungan produksi.

## Temuan 3: Validasi Format Email Manual dengan Regular Expression

**[Lokasi Kode]{.underline}**

> auth_controller.go :: Login()

**[Kode Sebelum Refactoring]{.underline}**

> emailRegex :=
> regexp.MustCompile(\`\^\[a-z0-9.\_%+\\-\]+@\[a-z0-9.\\-\]+\\.\[a-z\]{2,4}\$\`)
>
> if !emailRegex.MatchString(input.Email) { \... }

**[Masalah yang Ditemukan]{.underline}**

Menulis logika pencocokan pola regex secara manual di dalam handler
login mengotori alur pengecekan kredensial utama dan memicu duplikasi
jika fitur registrasi atau pembaruan profil membutuhkan validasi email
yang sama.

**[Prinsip Desain yang Dilanggar]{.underline}**

Clean Code (DRY - Don\'t Repeat Yourself): logika yang sama tidak boleh
diduplikasi di banyak tempat. Reusability: komponen validasi seharusnya
dapat digunakan ulang secara deklaratif.

**[Kode Sesudah Refactoring]{.underline}**

> // Memanfaatkan tag validator bawaan Gin Binding pada Struct DTO
>
> type LoginInput struct {
>
> Email string \`json:\"email\" binding:\"required,email\"\`
>
> Password string \`json:\"password\" binding:\"required\"\`
>
> }

**[Dampak Perbaikan]{.underline}**

Validasi email menjadi deklaratif dan terpusat. Tidak ada duplikasi
kode, dan penambahan aturan validasi cukup dilakukan dengan mengubah tag
pada struct DTO tanpa menyentuh logika handler.

## Temuan 4: Penurunan Performa Eksponensial N+1 Query Agregasi Rating

**[Lokasi Kode]{.underline}**

> umkm_controller.go :: GetProducts()

**[Kode Sebelum Refactoring]{.underline}**

> // Ambil semua produk (1 query)
>
> config.DB.Find(&allProducts)
>
> // Eksekusi N query tambahan di dalam perulangan
>
> for i := range allProducts {
>
> var stats struct { Average float64 }
>
> config.DB.Model(&models.Review{}).
>
> Where(\"product_id = ?\", allProducts\[i\].ID).Scan(&stats)
>
> allProducts\[i\].RatingAverage = stats.Average
>
> }

**[Masalah yang Ditemukan]{.underline}**

Mengambil seluruh baris produk lalu melakukan perulangan untuk
mengeksekusi kueri rating satu per satu memicu masalah N+1 Query. Jika
terdapat 100 produk, aplikasi melakukan 101 kueri ke database. Beban
basis data akan turun secara drastis saat diakses pada skala besar.

**[Prinsip Desain yang Dilanggar]{.underline}**

Database Efficiency & Performance Optimization: operasi database yang
mahal tidak boleh ditempatkan di dalam perulangan. Jumlah kueri harus
diminimalkan dengan mendelegasikan kalkulasi ke database engine.

**[Kode Sesudah Refactoring]{.underline}**

> // Menarik data beserta kalkulasi rata-rata dalam satu kueri Subquery
> tunggal
>
> config.DB.
>
> Select(\"products.\*, COALESCE((SELECT AVG(rating) FROM reviews\" +
>
> \" WHERE reviews.product_id = products.id), 0) as rating_average\").
>
> Find(&allProducts)

**[Dampak Perbaikan]{.underline}**

Beban database berkurang secara signifikan. Dari N+1 kueri menjadi hanya
1 kueri tunggal dengan kalkulasi AVG yang didelegasikan ke database
engine yang jauh lebih optimal.

## Temuan 5: Pemborosan Memori Validasi & Angka Magic Tarif Layanan

**[Lokasi Kode]{.underline}**

> umkm_controller.go :: CreateOrder()

**[Kode Sebelum Refactoring]{.underline}**

> // Menarik SELURUH ID produk ke RAM hanya untuk validasi satu item
>
> var sortedIDs \[\]string
>
> config.DB.Model(&models.Product{}).Pluck(\"id\", &sortedIDs)
>
> sort.Strings(sortedIDs)
>
> if !IsItemValid(sortedIDs, input.ItemID) { \... }
>
> // Tarif biaya layanan hardcoded langsung di rumus
>
> systemFee := totalBasePrice \* 0.03

**[Masalah yang Ditemukan]{.underline}**

Memuat seluruh ID produk dari SQL database ke memori RAM aplikasi hanya
untuk penanganan Binary Search lokal sangat memboroskan resource. Nilai
desimal 0.03 (tarif 3%) di-hardcode langsung dalam rumus matematis
transaksi.

**[Prinsip Desain yang Dilanggar]{.underline}**

Performance Optimization (Memory Bloat avoidance): validasi eksistensi
record seharusnya didelegasikan ke database engine. Clean Code (Magic
Value): konstanta bisnis tidak boleh dikodekan secara literal.

**[Kode Sesudah Refactoring]{.underline}**

> // Gunakan kueri indeks COUNT bawaan SQL untuk validasi
>
> var count int64
>
> config.DB.Model(&models.Product{}).Where(\"id = ?\",
> input.ItemID).Count(&count)
>
> if count == 0 { c.JSON(400, gin.H{\"error\": \"Item ID tidak
> valid\"}); return }
>
> // Tarif dipanggil dari konstanta global di config
>
> // config/app.go: const SupplierHubFeeRate = 0.03
>
> systemFee := totalBasePrice \* config.SupplierHubFeeRate

**[Dampak Perbaikan]{.underline}**

Penggunaan memori RAM menjadi konstan (O(1)) terlepas dari jumlah data
produk. Perubahan tarif platform di masa depan cukup dilakukan dengan
mengubah satu baris konstanta di config/app.go.

## Temuan 6: Penggabungan Logika Detail Algoritma KMP dalam Berkas Handler

**[Lokasi Kode]{.underline}**

> umkm_controller.go :: KMPMatch()

**[Kode Sebelum Refactoring]{.underline}**

> // Ditulis langsung di dalam file controller
>
> func KMPMatch(text string, pattern string) bool {
>
> // Perhitungan array tabel LPS (Longest Prefix Suffix)\...
>
> // Perulangan pencocokan string matematika\...
>
> }

**[Masalah yang Ditemukan]{.underline}**

Berkas pengontrol memiliki tanggung jawab ganda: mengurus alur HTTP REST
API dan mengurus detail matematis pergeseran indeks Knuth-Morris-Pratt.
Hal ini menurunkan nilai kohesi (Low Cohesion) dan membuat algoritma
tidak dapat digunakan ulang oleh modul lain.

**[Prinsip Desain yang Dilanggar]{.underline}**

Separation of Concerns (SoC): lapisan yang berbeda tidak boleh bercampur
dalam satu berkas. High Cohesion: setiap modul seharusnya memuat hanya
fungsi-fungsi yang memiliki keterkaitan kuat.

**[Kode Sesudah Refactoring]{.underline}**

> // Dipindahkan ke sub-paket utilitas murni
>
> // utils/algorithms.go -\> func KMPMatch(text, pattern string) bool
>
> // Controller hanya memanggil
>
> if utils.KMPMatch(product.Name, searchQuery) {
>
> filteredProducts = append(filteredProducts, product)
>
> }

**[Dampak Perbaikan]{.underline}**

Struktur file controller menjadi rapi. Fungsi KMP kini dapat digunakan
kembali (reusable) oleh modul admin atau supplier tanpa duplikasi kode.

## Temuan 7: High Coupling Terhadap Variabel Database Global config.DB

**[Lokasi Kode]{.underline}**

> supplier_controller.go :: CreateProduct()

**[Kode Sebelum Refactoring]{.underline}**

> // Handler mengakses config.DB global secara langsung
>
> if err := config.DB.Create(&newProduct).Error; err != nil { \... }

**[Masalah yang Ditemukan]{.underline}**

Ketergantungan langsung pada instansiasi global config.DB mengunci mati
pengontrol. Kode tidak dapat diuji secara otomatis (unit testing)
menggunakan Mock/Fake Database karena setiap pengujian akan selalu
memerlukan koneksi database asli.

**[Prinsip Desain yang Dilanggar]{.underline}**

SOLID (Dependency Inversion Principle): modul tingkat tinggi tidak boleh
bergantung langsung pada modul tingkat rendah. Keduanya harus bergantung
pada abstraksi. Low Coupling: komponen seharusnya dapat dipertukarkan.

**[Kode Sesudah Refactoring]{.underline}**

> // Menerapkan Design Pattern Repository dengan Dependency Injection
>
> type ProductRepository struct { db \*gorm.DB }
>
> func NewProductRepository(db \*gorm.DB) \*ProductRepository {
>
> return &ProductRepository{db: db}
>
> }
>
> func (r \*ProductRepository) Create(p \*models.Product) error {
>
> return r.db.Create(p).Error
>
> }

**[Dampak Perbaikan]{.underline}**

Derajat ikatan komponen menjadi rendah (Low Coupling). Handler dapat
diuji secara otomatis menggunakan Mock Repository tanpa memerlukan
koneksi database asli di setiap siklus pengujian.

## Temuan 8: Polusi Fungsi Pembentuk Payload JSON Respons Lokal

**[Lokasi Kode]{.underline}**

> supplier_controller.go :: supplierProfilePayload()

**[Kode Sebelum Refactoring]{.underline}**

> // Fungsi pembantu mengotori file controller
>
> func supplierProfilePayload(supplier models.User) gin.H {
>
> return gin.H{\"id\": supplier.ID, \"name\": supplier.BusinessName,
> \...}
>
> }

**[Masalah yang Ditemukan]{.underline}**

Menaruh fungsi pembantu transformasi struktur data map respons di luar
fungsionalitas utama berkas pengontrol mengurangi estetika kebersihan
arsitektur penyajian. File controller kehilangan fokus utamanya sebagai
HTTP handler.

**[Prinsip Desain yang Dilanggar]{.underline}**

Clean Code (Domain Cohesion): logika transformasi data dari entitas
domain ke format respons merupakan tanggung jawab model entitas, bukan
controller.

**[Kode Sesudah Refactoring]{.underline}**

> // Logika transformasi dienkapsulasi menjadi metode terikat milik
> model entitas
>
> // Di models/models.go
>
> func (u \*User) ToProfileResponse() gin.H {
>
> return gin.H{\"id\": u.ID, \"business_name\": u.BusinessName, \...}
>
> }
>
> // Controller menjadi bersih
>
> c.JSON(http.StatusOK, supplier.ToProfileResponse())

**[Dampak Perbaikan]{.underline}**

File controller bersih dari fungsi pembantu yang tidak relevan. Logika
pembentukan respons terisolasi dengan rapi di level domain model dan
dapat digunakan ulang.

## Temuan 9: Duplikasi Logika Normalisasi Teks Kondisi Status Pesanan

**[Lokasi Kode]{.underline}**

> supplier_controller.go :: UpdateOrderStatus()

**[Kode Sebelum Refactoring]{.underline}**

> // Normalisasi string ditulis berulang di beberapa tempat
>
> status := strings.ToLower(input.Status)
>
> if status == \"pending\" \|\| status == \"pnd\" { \... }

**[Masalah yang Ditemukan]{.underline}**

Melakukan manipulasi string penyeragaman status pesanan berulang kali di
level pengontrol rawan memicu celah inkonsistensi data jika tim
menambahkan status operasional baru di masa mendatang.

**[Prinsip Desain yang Dilanggar]{.underline}**

Clean Code (DRY - Don\'t Repeat Yourself): setiap logika bisnis
seharusnya memiliki representasi tunggal yang terpusat dalam sistem.

**[Kode Sesudah Refactoring]{.underline}**

> // Membungkus aturan transisi status ke dalam tipe data Enum-like di
> domain model
>
> // Di models/models.go
>
> type OrderStatus string
>
> func (o \*Order) TransitionTo(newStatus OrderStatus) error {
>
> // Logika validasi dan transisi status terpusat di sini
>
> return nil
>
> }

**[Dampak Perbaikan]{.underline}**

Logika normalisasi dan validasi transisi status pesanan terisolasi di
satu tempat. Penambahan status baru cukup dilakukan di model entitas
tanpa menyentuh controller manapun.

## Temuan 10: Penulisan Kueri Kompleks \'Like Search\' yang Mengotori Handler

**[Lokasi Kode]{.underline}**

> admin_controller.go :: GetAdminSuppliers()

**[Kode Sebelum Refactoring]{.underline}**

> // Rantai kueri panjang ditulis langsung di handler
>
> query = query.Where(
>
> \"business_name LIKE ? OR email LIKE ? OR category LIKE ? OR region
> LIKE ?\",
>
> likeSearch, likeSearch, likeSearch, likeSearch,
>
> )

**[Masalah yang Ditemukan]{.underline}**

Menuliskan gabungan operasi kueri string database OR yang panjang
langsung di rantai eksekusi GORM statement handler admin melanggar batas
pemisahan lapisan arsitektur (Presentation Layer).

**[Prinsip Desain yang Dilanggar]{.underline}**

Separation of Concerns / Layered Architecture: query logic harus
dipisahkan dari HTTP handling logic. Low Coupling: controller seharusnya
tidak mengetahui detail SQL secara langsung.

**[Kode Sesudah Refactoring]{.underline}**

> // Memanfaatkan fitur GORM Scopes untuk mengisolasi query di domain
> model
>
> // Di models/models.go
>
> func SearchSuppliersScope(term string) func(db \*gorm.DB) \*gorm.DB {
>
> return func(db \*gorm.DB) \*gorm.DB {
>
> like := \"%\" + term + \"%\"
>
> return db.Where(\"business_name LIKE ? OR email LIKE ?\", like, like)
>
> }
>
> }
>
> // Controller menjadi minimalis
>
> config.DB.Scopes(models.SearchSuppliersScope(searchQuery)).Find(&suppliers)

**[Dampak Perbaikan]{.underline}**

Logika query SQL yang kompleks terisolasi di tempat semestinya.
Controller menjadi sangat minimalis dan mudah dibaca sebagai alur
bisnis.

## Temuan 11: Penurunan Keterbacaan Akibat Inisialisasi Slice Array Lokal Panjang

**[Lokasi Kode]{.underline}**

> admin_controller.go :: GetAdminStats()

**[Kode Sebelum Refactoring]{.underline}**

> // Deklarasi panjang di dalam fungsi inti mengalihkan fokus
>
> activeStatuses := \[\]models.OrderStatus{
>
> models.OrderPending, models.OrderPendingSupplierConfirmation,
>
> models.OrderPaid, models.OrderProcessing
>
> }

**[Masalah yang Ditemukan]{.underline}**

Pendefinisian baris data array statis status pesanan aktif yang panjang
di dalam tubuh fungsi utama GetAdminStats merusak fokus keterbacaan
baris alur agregasi statistik utama.

**[Prinsip Desain yang Dilanggar]{.underline}**

Clean Code (Readability Improvement): kode yang mudah dibaca adalah kode
yang mengungkapkan niatnya dengan jelas. Deklarasi data statis yang
panjang mengotori alur logika utama.

**[Kode Sesudah Refactoring]{.underline}**

> // Dipindahkan menjadi variabel global milik paket model entitas
>
> // Di models/models.go
>
> var ActiveOrderStatuses = \[\]OrderStatus{
>
> OrderPending, OrderPaid, OrderProcessing, OrderShipped
>
> }
>
> // Handler menjadi lebih fokus
>
> config.DB.Where(\"status IN ?\",
> models.ActiveOrderStatuses).Count(&total)

**[Dampak Perbaikan]{.underline}**

Baris kode fungsi statistik admin berkurang signifikan. Pembaca kode
dapat langsung memahami alur logika agregasi tanpa terganggu oleh
deklarasi data statis yang panjang.

## Temuan 12: Low Cohesion pada Fungsi Konversi Utilitas Generik toString

**[Lokasi Kode]{.underline}**

> admin_controller.go :: toString()

**[Kode Sebelum Refactoring]{.underline}**

> // Ditemukan di bagian bawah file admin_controller.go
>
> func toString(value interface{}) string { \... }

**[Masalah yang Ditemukan]{.underline}**

Fungsi toString bertugas mengubah interface mentah menjadi string
generik. Menanam fungsi pembantu tipe data dasar ini di dalam berkas
administratif admin sangat menyalahi struktur kohesi modul karena tidak
ada kaitannya dengan domain bisnis administratif.

**[Prinsip Desain yang Dilanggar]{.underline}**

High Cohesion: sebuah modul seharusnya memuat hanya fungsi-fungsi yang
memiliki keterkaitan kuat satu sama lain dengan domain bisnis yang sama.

**[Kode Sesudah Refactoring]{.underline}**

> // Dipindahkan ke berkas helper independen yang dapat dipakai berkas
> lain
>
> // Di utils/converter.go
>
> package utils
>
> func InterfaceToString(val interface{}) string {
>
> // Implementasi konversi tipe data
>
> }

**[Dampak Perbaikan]{.underline}**

Nilai kohesi controller admin meningkat karena hanya memuat fungsi yang
relevan dengan domain administrasi. Fungsi konversi kini tersedia untuk
seluruh modul sebagai utilitas reusable.

## Temuan 13: Instansiasi Langsung HTTP Client Menghalangi Mocking Testing Eksternal API

**[Lokasi Kode]{.underline}**

> payment_controller.go :: forwardPaymentToSmartBank()

**[Kode Sebelum Refactoring]{.underline}**

> // HTTP Client dibuat langsung di dalam fungsi
>
> client := &http.Client{Timeout: 10 \* time.Second}
>
> resp, err := client.Do(req)

**[Masalah yang Ditemukan]{.underline}**

Membuat objek instansiasi HTTP client bawaan secara langsung mengunci
mati kode transaksi untuk selalu memukul network server rill gateway
SmartBank. Proses simulasi uji coba kegagalan jaringan internal menjadi
mustahil dijalankan.

**[Prinsip Desain yang Dilanggar]{.underline}**

SOLID (Dependency Inversion Principle): fungsi seharusnya bergantung
pada abstraksi (interface), bukan implementasi konkret. Testability:
komponen eksternal harus dapat dipertukarkan dengan mock saat testing.

**[Kode Sesudah Refactoring]{.underline}**

> // Menyediakan interface abstraksi HTTP client
>
> type HTTPClientInterface interface {
>
> Do(req \*http.Request) (\*http.Response, error)
>
> }
>
> // Handler menerima client melalui dependency injection
>
> type PaymentHandler struct { httpClient HTTPClientInterface }
>
> // Saat testing: inject MockHTTPClient
>
> // Saat produksi: inject &http.Client{Timeout: 10\*time.Second}

**[Dampak Perbaikan]{.underline}**

Pengujian simulasi skenario kegagalan koneksi SmartBank kini dapat
dilakukan dengan aman menggunakan mock client tanpa menyentuh server
rill di setiap siklus pengujian otomatis.

## Temuan 14: Hardcoded Nilai Nomor Port Web Server Internal

**[Lokasi Kode]{.underline}**

> main.go :: main()

**[Kode Sebelum Refactoring]{.underline}**

> // Port dikunci secara statis
>
> r.Run(\":8080\")

**[Masalah yang Ditemukan]{.underline}**

Mengunci string nomor port server secara kaku pada nilai statis
\":8080\" membuat aplikasi tidak dapat dijalankan secara fleksibel di
infrastruktur cloud modern (seperti Docker, Cloud Run, Heroku) yang
mengalokasikan port secara dinamis melalui variabel lingkungan sistem
operasi.

**[Prinsip Desain yang Dilanggar]{.underline}**

Clean Code / Twelve-Factor App Deployment: konfigurasi deployment harus
dapat diubah dari lingkungan (environment) tanpa modifikasi kode sumber.

**[Kode Sesudah Refactoring]{.underline}**

> port := os.Getenv(\"PORT\")
>
> if port == \"\" { port = \"8080\" } // Fallback aman untuk local dev
>
> r.Run(\":\" + port)

**[Dampak Perbaikan]{.underline}**

Aplikasi kini fleksibel dijalankan di server lokal maupun platform cloud
(Docker, Kubernetes, Cloud Run) yang mengalokasikan port secara dinamis
melalui ENV.

## Temuan 15: Polusi Sinkronisasi State Sesi Token Dualistik

**[Lokasi Kode]{.underline}**

> Login/login.html & assets/js/auth.js

**[Kode Sebelum Refactoring]{.underline}**

> // Menyimpan di key \"token\"
>
> localStorage.setItem(\"token\", result.token);
>
> // Di bagian script lain mengakses key berbeda
>
> const session = JSON.parse(localStorage.getItem(\"user_session\"));

**[Masalah yang Ditemukan]{.underline}**

Terjadi pembagian state management yang membingungkan (dualistic token
state). Sebagian komponen frontend menyimpan nilai token di key terpisah
(\"token\"), sementara modul dashboard memuat objek terstruktur
\"user_session\". Hal ini mengakibatkan risiko bug autentikasi fatal.

**[Prinsip Desain yang Dilanggar]{.underline}**

State Integrity / Shared Session Helper: seluruh mutasi data sesi
seharusnya dilakukan melalui satu helper tunggal yang terstandarisasi
agar tidak terjadi inkonsistensi antar komponen.

**[Kode Sesudah Refactoring]{.underline}**

> // Memusatkan seluruh mutasi data sesi pada helper tunggal di auth.js
>
> function saveUserSession(authResponse) {
>
> // Selalu menulis ke key yang sama: \"user_session\"
>
> localStorage.setItem(\"user_session\", JSON.stringify(authResponse));
>
> }
>
> // Dipanggil konsisten dari semua titik login
>
> saveUserSession(result); // Konsisten menulis ke
> localStorage.user_session

**[Dampak Perbaikan]{.underline}**

Seluruh komponen frontend kini membaca dan menulis sesi dari satu key
yang sama. Risiko bug autentikasi akibat inkonsistensi key localStorage
dieliminasi sepenuhnya.

# 6. Arsitektur Komponen Setelah Perbaikan (Graphviz DOT)

Berikut rancangan relasi arsitektur aplikasi SupplierHub setelah
dibersihkan dari ikatan kaku objek basis data global (config.DB) dan
penataan fungsi-fungsi gemuk pengontrol.

> digraph SupplierHubCleanArchitecture {
>
> rankdir=LR;
>
> node \[shape=record, fontname=\"Arial\", style=\"filled\"\];
>
> subgraph cluster_presentation_layer {
>
> label = \"PRESENTATION LAYER (Front-End Statis)\";
>
> UI \[label=\"{Vanilla JS Fetch API \| + Update Navbar\\l+ Render List
> DTO\\l}\"\];
>
> }
>
> subgraph cluster_http_delivery {
>
> label = \"HTTP DELIVERY LAYER (Go Gin)\";
>
> Handlers \[label=\"{Controllers (Handlers) \| + Register(c)\\l+
> CreateOrder(c)\\l+ GetAdminStats(c)\\l}\"\];
>
> }
>
> subgraph cluster_business_logic {
>
> label = \"BUSINESS LOGIC LAYER\";
>
> Services \[label=\"{Domain Services \| + UserService (Hash & Save)\\l+
> OrderService (Fee 3%)\\l}\"\];
>
> Algorithms \[label=\"{utils/algorithms \| + KMPMatch()\\l+
> QuickSortPrice()\\l+ IsItemValid()\\l}\"\];
>
> }
>
> subgraph cluster_data_access {
>
> label = \"DATA ACCESS LAYER\";
>
> Repositories \[label=\"{Repositories (Injected DB) \| +
> UserRepository\\l+ ProductRepository\\l}\"\];
>
> DB \[label=\"{MySQL Database \| + AutoMigrate Schema}\",
> shape=cylinder\];
>
> }
>
> UI -\> Handlers \[label=\" HTTP REST Requests\"\];
>
> Handlers -\> Services \[label=\" Invoke Use Case DTO\"\];
>
> Services -\> Repositories \[label=\" Database Operations\"\];
>
> Services -\> Algorithms \[label=\" Delegate Core Comp.\",
> style=\"dashed\"\];
>
> Repositories -\> DB \[label=\" Read/Write SQL\"\];
>
> }

# 7. Kesimpulan Refactoring

Berdasarkan analisis kode, SupplierHub sudah menggunakan struktur MVC
sederhana yang memisahkan controller, model, route, dan services.
Aplikasi juga memiliki fitur bisnis nyata seperti algoritma KMP, Quick
Sort, Binary Search, dan integrasi payment gateway SmartBank.

Melalui pengklasteran dan eksekusi 15 perbaikan terstruktur ini,
codebase SupplierHub bertransformasi menjadi sistem yang berkohesi
tinggi (High Cohesion) dan berderajat ikatan rendah (Low Coupling).
Secara spesifik:

1.  Kode program kini sepenuhnya bersih dari polusi nilai mentah (Magic
    Values) yang sebelumnya tersebar di berbagai file controller.

2.  Aplikasi aman dari celah kebocoran token di tingkat produksi berkat
    penghapusan fallback JWT secret hardcoded.

3.  Performa server meningkat signifikan berkat eliminasi N+1 Query
    Problem pada agregasi rating produk.

4.  Kode siap dilakukan pengujian otomatis (testable) secara end-to-end
    berkat penerapan pola Repository dengan Dependency Injection.

5.  Algoritma inti (KMP, Quick Sort, Binary Search) kini terisolasi di
    paket utils yang reusable dan bebas dari ketergantungan HTTP
    context.

# 8. Lampiran

## 8.1 Link Repository

> https://github.com/sprint-projcet/SUPPLIER-HUB.git

## 8.2 Branch Sebelum dan Sesudah Refactoring

  ---------------------------------------------------------------------------------------------------
  **Jenis Branch**                   **Nama Branch**
  ---------------------------------- ----------------------------------------------------------------
  **Branch sebelum refactoring**     https://github.com/sprint-projcet/SUPPLIER-HUB.git

  **Branch sesudah refactoring**     https://github.com/sprint-projcet/SUPPLIER-HUB-REFACTORING.git
  ---------------------------------------------------------------------------------------------------

## 8.3 Rekomendasi Struktur Folder Setelah Refactoring

> SUPPLIER-HUB/backend/
>
> ├── config/
>
> │ ├── app.go // Konstanta global (FeeRate, UploadDir, dll)
>
> │ └── database.go
>
> ├── controllers/ // Hanya HTTP handler (ramping)
>
> ├── models/ // Struct + method domain (ToProfileResponse,
> TransitionTo)
>
> ├── repositories/ // (BARU) Data Access Object dengan Dependency
> Injection
>
> │ ├── product_repository.go
>
> │ └── user_repository.go
>
> ├── services/ // (DIPERLUAS) Business logic layer
>
> │ ├── user_service.go
>
> │ └── order_service.go
>
> ├── utils/ // (BARU) Cross-cutting helpers
>
> │ ├── algorithms.go // KMPMatch, QuickSort, BinarySearch
>
> │ └── converter.go // InterfaceToString, dll
>
> ├── routes/
>
> └── main.go
