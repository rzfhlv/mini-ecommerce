package main

import (
	"log"
	"net/http"
	"time"

	catalogdelivery "mini-ecommerce/internal/catalog/delivery"
	catalogadapter "mini-ecommerce/internal/catalog/infrastructure"
	catalogpublic "mini-ecommerce/internal/catalog/public"
	catalogusecase "mini-ecommerce/internal/catalog/usecase"

	cartdelivery "mini-ecommerce/internal/cart/delivery"
	cartinfra "mini-ecommerce/internal/cart/infrastructure"
	cartusecase "mini-ecommerce/internal/cart/usecase"

	orderdelivery "mini-ecommerce/internal/order/delivery"
	orderdomain "mini-ecommerce/internal/order/domain"
	orderinfra "mini-ecommerce/internal/order/infrastructure"
	orderusecase "mini-ecommerce/internal/order/usecase"

	userdelivery "mini-ecommerce/internal/user/delivery"
	userdomain "mini-ecommerce/internal/user/domain"
	userinfra "mini-ecommerce/internal/user/infrastructure"
	userusecase "mini-ecommerce/internal/user/usecase"

	"mini-ecommerce/pkg/config"
	"mini-ecommerce/pkg/database"
	"mini-ecommerce/pkg/events"
	"mini-ecommerce/pkg/hash"
	"mini-ecommerce/pkg/jwt"
	"mini-ecommerce/pkg/middleware"
)

// ============================================================
// COMPOSITION ROOT
// ============================================================
// EMPAT bounded context: user, catalog, cart, order. Urutan wiring
// PENTING: catalog duluan (order & cart butuh CatalogPublicAPI-nya),
// baru cart & order. user independen, cuma "bertemu" context lain
// lewat middleware.Auth (inject user_id ke context).
// ============================================================

func main() {
	cfg := config.Load()

	db, err := database.NewPostgresConnection(database.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser,
		Password: cfg.DBPassword, DBName: cfg.DBName, SSLMode: cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

	// ---------- Shared ----------
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret, 24*time.Hour)
	passwordHasher := hash.NewBcryptHasher()
	authMiddleware := middleware.Auth(tokenManager)

	// ---------- User ----------
	userRepo := userinfra.NewPostgresUserRepository(db)
	registrationService := userdomain.NewRegistrationService(userRepo, passwordHasher)
	authService := userdomain.NewAuthenticationService(userRepo, passwordHasher)
	registerUC := userusecase.NewRegisterUseCase(registrationService)
	loginUC := userusecase.NewLoginUseCase(authService, tokenManager)
	getMeUC := userusecase.NewGetMeUseCase(userRepo)
	userHandler := userdelivery.NewUserHandler(registerUC, loginUC, getMeUC)

	// ---------- Catalog ----------
	productRepo := catalogadapter.NewPostgresProductRepository(db)
	categoryRepo := catalogadapter.NewPostgresCategoryRepository(db)
	catalogAPI := catalogpublic.NewCatalogPublicAPI(productRepo)

	createProductUC := catalogusecase.NewCreateProductUseCase(productRepo)
	updateProductUC := catalogusecase.NewUpdateProductUseCase(productRepo)
	deleteProductUC := catalogusecase.NewDeleteProductUseCase(productRepo)
	getProductUC := catalogusecase.NewGetProductUseCase(productRepo)
	listProductsUC := catalogusecase.NewListProductsUseCase(productRepo)
	listCategoriesUC := catalogusecase.NewListCategoriesUseCase(categoryRepo)

	productHandler := catalogdelivery.NewProductHandler(createProductUC, updateProductUC, deleteProductUC, getProductUC, listProductsUC)
	categoryHandler := catalogdelivery.NewCategoryHandler(listCategoriesUC)

	// ---------- Cart ----------
	cartRepo := cartinfra.NewPostgresCartRepository(db)
	cartPricer := cartinfra.NewCatalogPricerAdapter(catalogAPI)
	addItemUC := cartusecase.NewAddItemUseCase(cartRepo, cartPricer)
	updateItemQtyUC := cartusecase.NewUpdateItemQuantityUseCase(cartRepo)
	removeItemUC := cartusecase.NewRemoveItemUseCase(cartRepo)
	getCartUC := cartusecase.NewGetCartUseCase(cartRepo)
	clearCartUC := cartusecase.NewClearCartUseCase(cartRepo)
	cartHandler := cartdelivery.NewCartHandler(addItemUC, updateItemQtyUC, removeItemUC, getCartUC, clearCartUC)

	// ---------- Order ----------
	orderRepo := orderinfra.NewPostgresOrderRepository(db)
	stockChecker := orderinfra.NewCatalogStockCheckerAdapter(catalogAPI)
	promoRepo := orderinfra.NewPostgresPromoCodeRepository(db)
	promoAdminRepo := orderinfra.NewPostgresPromoAdminRepository(db)
	paymentGateway := orderinfra.NewMidtransPaymentGateway(cfg.MidtransBaseURL, cfg.MidtransServerKey)
	eventPublisher := events.NewLogEventPublisher()

	// CheckoutService sekarang PURE -- cuma butuh stockChecker.
	checkoutService := orderdomain.NewCheckoutService(stockChecker)
	// CheckoutUseCase yang mengorkestrasi SEMUA efek samping tulis
	// (ReduceStock, Save, payment gateway) -- lihat refactor Pendekatan B.
	checkoutUC := orderusecase.NewCheckoutUseCase(checkoutService, stockChecker, orderRepo, promoRepo, paymentGateway, eventPublisher)
	confirmPaymentUC := orderusecase.NewConfirmPaymentUseCase(orderRepo, eventPublisher)
	orderHandler := orderdelivery.NewOrderHandler(checkoutUC, confirmPaymentUC, paymentGateway)

	createPromoUC := orderusecase.NewCreatePromoUseCase(promoAdminRepo)
	updatePromoUC := orderusecase.NewUpdatePromoUseCase(promoAdminRepo)
	setPromoActiveUC := orderusecase.NewSetPromoActiveUseCase(promoAdminRepo)
	deletePromoUC := orderusecase.NewDeletePromoUseCase(promoAdminRepo)
	getPromoUC := orderusecase.NewGetPromoUseCase(promoAdminRepo)
	listPromoUC := orderusecase.NewListPromoUseCase(promoAdminRepo)
	promoAdminHandler := orderdelivery.NewPromoAdminHandler(createPromoUC, updatePromoUC, setPromoActiveUC, deletePromoUC, getPromoUC, listPromoUC)

	// ---------- Routing ----------
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", userHandler.Login)
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/v1/products", productHandler.List)
	mux.HandleFunc("GET /api/v1/products/{id}", productHandler.Get)
	mux.HandleFunc("POST /api/v1/webhooks/payment", orderHandler.PaymentWebhook)
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	mux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(userHandler.Me)))
	mux.Handle("POST /api/v1/orders/checkout", authMiddleware(http.HandlerFunc(orderHandler.Checkout)))

	mux.Handle("POST /api/v1/admin/products", authMiddleware(http.HandlerFunc(productHandler.Create)))
	mux.Handle("PUT /api/v1/admin/products/{id}", authMiddleware(http.HandlerFunc(productHandler.Update)))
	mux.Handle("DELETE /api/v1/admin/products/{id}", authMiddleware(http.HandlerFunc(productHandler.Delete)))

	mux.Handle("POST /api/v1/admin/promos", authMiddleware(http.HandlerFunc(promoAdminHandler.Create)))
	mux.Handle("GET /api/v1/admin/promos", authMiddleware(http.HandlerFunc(promoAdminHandler.List)))
	mux.Handle("GET /api/v1/admin/promos/{id}", authMiddleware(http.HandlerFunc(promoAdminHandler.Get)))
	mux.Handle("PUT /api/v1/admin/promos/{id}", authMiddleware(http.HandlerFunc(promoAdminHandler.Update)))
	mux.Handle("DELETE /api/v1/admin/promos/{id}", authMiddleware(http.HandlerFunc(promoAdminHandler.Delete)))
	mux.Handle("PATCH /api/v1/admin/promos/{id}/activate", authMiddleware(http.HandlerFunc(promoAdminHandler.Activate)))
	mux.Handle("PATCH /api/v1/admin/promos/{id}/deactivate", authMiddleware(http.HandlerFunc(promoAdminHandler.Deactivate)))

	mux.Handle("GET /api/v1/cart", authMiddleware(http.HandlerFunc(cartHandler.Get)))
	mux.Handle("DELETE /api/v1/cart", authMiddleware(http.HandlerFunc(cartHandler.Clear)))
	mux.Handle("POST /api/v1/cart/items", authMiddleware(http.HandlerFunc(cartHandler.AddItem)))
	mux.Handle("PUT /api/v1/cart/items/{productId}", authMiddleware(http.HandlerFunc(cartHandler.UpdateItem)))
	mux.Handle("DELETE /api/v1/cart/items/{productId}", authMiddleware(http.HandlerFunc(cartHandler.RemoveItem)))

	log.Printf("server listening on :%s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, mux))
}
