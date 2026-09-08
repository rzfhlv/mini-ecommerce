# Mini E-Commerce Backend — DDD Full Example

Clean Architecture + DDD, 4 Bounded Context: **User**, **Catalog**, **Cart**, **Order**
(termasuk Discount/Promo, Admin Fee, dan Payment ACL).

## Struktur

```
cmd/api/main.go        Composition Root

internal/
  user/                 Identity/Auth (Generic Domain)
  catalog/               Product & Category (Core/Supporting)
    public/                Open Host Service -- satu-satunya pintu context lain
  cart/                  Keranjang belanja (Supporting, independen dari Order)
  order/                 Checkout, Discount, Admin Fee, Payment (Core)

pkg/                    shared infra: jwt, hash, database, config, middleware, events
migrations/             SQL schema
```

## Fitur lengkap di module Order

- **Discount** via Strategy Pattern (`DiscountPolicy`): persentase, nominal
  tetap, syarat minimum pembelian (decorator), Null Object untuk "tanpa diskon".
- **Promo Code CRUD** oleh admin (`PromoCodeAdminRepository`), terpisah dari
  `PromoCodeRepository` yang dipakai proses checkout.
- **Admin Fee** (BARU) via Strategy Pattern terpisah (`AdminFeePolicy`):
  biaya tetap (transfer bank), persentase basis-points (kartu kredit 2.9%),
  gratis (e-wallet). Ditentukan dari `PaymentMethod`.
- **Payment** via Anti-Corruption Layer (`MidtransPaymentGateway`) -- BUKAN
  bounded context penuh, karena Payment diklasifikasikan Generic Domain.
- **CheckoutService PURE** (refactor terbaru): tidak lagi memanggil
  `OrderRepository.Save()` atau `ReduceStock()` -- itu sekarang tanggung
  jawab `CheckoutUseCase`. Lihat komentar di `checkout_service.go` dan
  `checkout_usecase.go` untuk penjelasan trade-off Pendekatan A vs B.

## Aturan boundary antar Bounded Context

```bash
# Order/Cart cuma boleh sentuh catalog/public, tidak pernah catalog/domain
grep -rln "internal/catalog" internal/order/ | xargs grep -L "catalog/public"
grep -rln "internal/catalog" internal/cart/ | xargs grep -L "catalog/public"
# Catalog tidak boleh tahu Order/Cart/User sama sekali
grep -rln "internal/order\|internal/cart\|internal/user" internal/catalog/
# Order dan Cart independen total, tidak saling import
grep -rln "internal/cart" internal/order/
grep -rln "internal/order" internal/cart/
```
Semua harus mengembalikan hasil kosong.

## Menjalankan

```bash
cp .env.example .env
docker-compose up --build
```

## Contoh alur

```bash
curl -X POST localhost:8080/api/v1/auth/register -d '{"name":"Budi","email":"budi@example.com","password":"secret123"}'
curl -X POST localhost:8080/api/v1/auth/login -d '{"email":"budi@example.com","password":"secret123"}'

curl localhost:8080/api/v1/products?category=shirt

curl -X POST localhost:8080/api/v1/cart/items -H "Authorization: Bearer <token>" -d '{"product_id":"<id>","quantity":2}'

curl -X POST localhost:8080/api/v1/orders/checkout -H "Authorization: Bearer <token>" \
  -d '{"items":[{"product_id":"<id>","quantity":2}],"promo_code":"HEMAT10","payment_method":"CREDIT_CARD"}'
```

## Yang belum diimplementasikan

- Module role/permission (endpoint `/admin/*` masih `authMiddleware` biasa)
- Signature verification di webhook pembayaran
- Validasi request body pakai `go-playground/validator`
