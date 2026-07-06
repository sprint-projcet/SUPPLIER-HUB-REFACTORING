package controllers

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"supplierhub-backend/config"
	"supplierhub-backend/models"
	"supplierhub-backend/repositories"
	"supplierhub-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type updateSupplierOrderStatusInput struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status" binding:"required"`
}

type supplierProfileInput struct {
	BusinessName string `json:"business_name" binding:"required"`
	Address      string `json:"address" binding:"required"`
	Category     string `json:"category" binding:"required"`
	Region       string `json:"region" binding:"required"`
}

// supplierProfilePayload has been moved to models.User.ToProfileResponse()

func getCurrentSupplier(c *gin.Context) (models.User, bool) {
	var supplier models.User

	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return supplier, false
	}

	if err := config.DB.Where("id = ? AND role = ?", supplierID, models.RoleSupplier).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profil supplier tidak ditemukan"})
			return supplier, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil profil supplier"})
		return supplier, false
	}

	return supplier, true
}

func GetSupplierProfile(c *gin.Context) {
	supplier, ok := getCurrentSupplier(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   supplier.ToProfileResponse(),
	})
}

func UpdateSupplierProfile(c *gin.Context) {
	supplier, ok := getCurrentSupplier(c)
	if !ok {
		return
	}

	var input supplierProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama toko, kategori, wilayah, dan alamat wajib diisi"})
		return
	}

	businessName := strings.TrimSpace(input.BusinessName)
	category := strings.TrimSpace(input.Category)
	region := strings.TrimSpace(input.Region)
	address := strings.TrimSpace(input.Address)
	if businessName == "" || category == "" || region == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama toko, kategori, wilayah, dan alamat wajib diisi"})
		return
	}

	updates := map[string]interface{}{
		"business_name": businessName,
		"address":       address,
		"category":      category,
		"region":        region,
	}

	if err := config.DB.Model(&supplier).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui profil supplier"})
		return
	}

	if err := config.DB.Where("id = ?", supplier.ID).First(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat ulang profil supplier"})
		return
	}

	description := "Profil toko " + supplier.BusinessName + " diperbarui"
	_ = services.CreateActivityLog(nil, supplier.ID, "UPDATE_SUPPLIER_PROFILE", description)
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     supplier.ID,
		Role:       string(models.RoleSupplier),
		Title:      "Profil Toko Diperbarui",
		Message:    description,
		Type:       "supplier_profile_updated",
		SourceType: "supplier",
		SourceID:   supplier.ID,
	})
	_, _ = services.CreateRoleActivityNotifications(nil, models.RoleAdmin, models.Notification{
		Title:      "Profil Supplier Diperbarui",
		Message:    description,
		Type:       "supplier_profile_updated",
		SourceType: "supplier",
		SourceID:   supplier.ID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profil toko berhasil diperbarui",
		"data":    supplier.ToProfileResponse(),
	})
}

func GetSupplierStats(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var completedOrders int64
	var pendingOrders int64
	var activeProducts int64
	var totalStock int64
	var supplierRevenue float64
	pendingStatuses := []models.OrderStatus{
		models.OrderPending,
		models.OrderPendingSupplierConfirmation,
		models.OrderSupplierConfirmed,
		models.OrderPaymentPending,
	}
	revenueStatuses := []models.OrderStatus{
		models.OrderPaid,
		models.OrderShipmentCreated,
		models.OrderCompleted,
	}

	if err := config.DB.Model(&models.Order{}).
		Where("supplier_id = ? AND status = ?", supplierID, models.OrderCompleted).
		Count(&completedOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung pesanan selesai"})
		return
	}

	if err := config.DB.Model(&models.Order{}).
		Where("supplier_id = ? AND status IN ?", supplierID, pendingStatuses).
		Count(&pendingOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung pesanan pending"})
		return
	}

	if err := config.DB.Model(&models.Product{}).
		Where("supplier_id = ? AND stock > 0", supplierID).
		Count(&activeProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung produk aktif"})
		return
	}

	if err := config.DB.Model(&models.Product{}).
		Where("supplier_id = ?", supplierID).
		Select("COALESCE(SUM(stock), 0)").
		Scan(&totalStock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung total stok"})
		return
	}

	if err := config.DB.Model(&models.Order{}).
		Where("supplier_id = ? AND status IN ?", supplierID, revenueStatuses).
		Select("COALESCE(SUM(total_base_price - system_fee), 0)").
		Scan(&supplierRevenue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung pendapatan supplier"})
		return
	}

	// Hitung rata-rata rating nyata milik supplier ini
	var avgRating float64
	config.DB.Model(&models.Review{}).
		Joins("JOIN products ON reviews.product_id = products.id").
		Where("products.supplier_id = ?", supplierID).
		Select("COALESCE(AVG(reviews.rating), 0)").
		Scan(&avgRating)

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"total_revenue":    supplierRevenue,
		"completed_orders": completedOrders,
		"active_products":  activeProducts,
		"pending_orders":   pendingOrders,
		"total_stock":      totalStock,
		"stock":            activeProducts,  // kompatibel dengan dashboard lama
		"new_orders":       pendingOrders,   // kompatibel dengan dashboard lama
		"revenue_rp":       supplierRevenue, // kompatibel dengan dashboard lama
		"rating":           avgRating,
	})
}

func GetSupplierProducts(c *gin.Context) {
	supplierID, _ := c.Get("user_id")

	var products []models.Product
	if err := config.DB.Where("supplier_id = ?", supplierID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func CreateProduct(c *gin.Context) {
	supplier, ok := getCurrentSupplier(c)
	if !ok {
		return
	}

	if supplier.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun supplier Anda belum aktif atau belum diverifikasi oleh admin."})
		return
	}

	supplierRegion := strings.TrimSpace(supplier.Region)
	if supplierRegion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lengkapi wilayah supply toko terlebih dahulu sebelum menambahkan produk."})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	category := strings.TrimSpace(c.PostForm("category"))
	priceStr := strings.TrimSpace(c.PostForm("price"))
	stockStr := strings.TrimSpace(c.PostForm("stock"))
	description := strings.TrimSpace(c.PostForm("description"))

	if name == "" || category == "" || priceStr == "" || stockStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama, kategori, harga, dan stok produk wajib diisi"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Harga produk tidak valid"})
		return
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil || stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stok produk tidak valid"})
		return
	}

	// File Upload Handling
	file, err := c.FormFile("image")
	var imageURL string
	if err == nil {
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		uploadPath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			imageURL = config.PublicURL(uploadPath)
		}
	}

	input := models.Product{
		SupplierID:  supplier.ID,
		Name:        name,
		Category:    category,
		Price:       price,
		Stock:       stock,
		Description: description,
		Location:    supplierRegion,
		ImageURL:    imageURL,
	}

	if err := repositories.NewProductRepository(config.DB).Create(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk: " + err.Error()})
		return
	}

	activityDescription := "Supplier " + supplier.BusinessName + " menambahkan produk " + input.Name + " dengan stok " + strconv.Itoa(input.Stock) + " unit"
	_ = services.CreateActivityLog(nil, supplier.ID, "CREATE_PRODUCT", activityDescription)
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     supplier.ID,
		Role:       string(models.RoleSupplier),
		Title:      "Produk Ditambahkan",
		Message:    activityDescription,
		Type:       "product_created",
		SourceType: "product",
		SourceID:   input.ID,
	})
	_, _ = services.CreateRoleActivityNotifications(nil, models.RoleAdmin, models.Notification{
		Title:      "Produk Supplier Ditambahkan",
		Message:    activityDescription,
		Type:       "product_created",
		SourceType: "product",
		SourceID:   input.ID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil ditambahkan",
		"data":    input,
	})
}

func UpdateProduct(c *gin.Context) {
	supplier, ok := getCurrentSupplier(c)
	if !ok {
		return
	}

	if supplier.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun supplier Anda belum aktif atau belum diverifikasi oleh admin."})
		return
	}

	supplierRegion := strings.TrimSpace(supplier.Region)
	if supplierRegion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lengkapi wilayah supply toko terlebih dahulu sebelum mengubah produk."})
		return
	}

	productID := c.Param("id")

	var product models.Product
	if err := config.DB.Where("id = ? AND supplier_id = ?", productID, supplier.ID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan atau Anda tidak berwenang"})
		return
	}
	oldStock := product.Stock
	oldName := product.Name

	name := c.PostForm("name")
	category := c.PostForm("category")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	description := c.PostForm("description")
	stockChanged := false

	if name != "" {
		product.Name = strings.TrimSpace(name)
	}
	if category != "" {
		product.Category = strings.TrimSpace(category)
	}
	if priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			product.Price = price
		}
	}
	if stockStr != "" {
		stock, err := strconv.Atoi(stockStr)
		if err != nil || stock < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Stok produk tidak valid"})
			return
		}
		stockChanged = stock != product.Stock
		product.Stock = stock
	}
	if description != "" {
		product.Description = strings.TrimSpace(description)
	}
	product.Location = supplierRegion

	// File Upload Handling
	file, err := c.FormFile("image")
	if err == nil {
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		uploadPath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			product.ImageURL = config.PublicURL(uploadPath)
		}
	}

	if err := config.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui produk: " + err.Error()})
		return
	}

	activityDescription := "Supplier " + supplier.BusinessName + " memperbarui produk " + oldName + " menjadi " + product.Name
	if stockChanged {
		activityDescription += ". Stok berubah dari " + strconv.Itoa(oldStock) + " menjadi " + strconv.Itoa(product.Stock) + " unit"
	}
	activityType := "product_updated"
	if stockChanged {
		activityType = "stock_updated"
	}
	_ = services.CreateActivityLog(nil, supplier.ID, strings.ToUpper(activityType), activityDescription)
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     supplier.ID,
		Role:       string(models.RoleSupplier),
		Title:      "Produk Diperbarui",
		Message:    activityDescription,
		Type:       activityType,
		SourceType: "product",
		SourceID:   product.ID,
	})
	_, _ = services.CreateRoleActivityNotifications(nil, models.RoleAdmin, models.Notification{
		Title:      "Produk Supplier Diperbarui",
		Message:    activityDescription,
		Type:       activityType,
		SourceType: "product",
		SourceID:   product.ID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diupdate", "data": product})
}

func DeleteProduct(c *gin.Context) {
	supplier, ok := getCurrentSupplier(c)
	if !ok {
		return
	}

	productID := c.Param("id")
	var product models.Product
	if err := config.DB.Where("id = ? AND supplier_id = ?", productID, supplier.ID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan atau Anda tidak berwenang"})
		return
	}

	if err := config.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk: " + err.Error()})
		return
	}

	description := "Supplier " + supplier.BusinessName + " menghapus produk " + product.Name
	_ = services.CreateActivityLog(nil, supplier.ID, "DELETE_PRODUCT", description)
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     supplier.ID,
		Role:       string(models.RoleSupplier),
		Title:      "Produk Dihapus",
		Message:    description,
		Type:       "product_deleted",
		SourceType: "product",
		SourceID:   product.ID,
	})
	_, _ = services.CreateRoleActivityNotifications(nil, models.RoleAdmin, models.Notification{
		Title:      "Produk Supplier Dihapus",
		Message:    description,
		Type:       "product_deleted",
		SourceType: "product",
		SourceID:   product.ID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil dihapus",
	})
}

func GetSupplierNotifications(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	query := config.DB.Where("user_id = ? AND role = ?", supplierID, string(models.RoleSupplier))
	if c.Query("unread_only") == "true" {
		query = query.Where("is_read = ?", false)
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("read_status"))) {
	case "read":
		query = query.Where("is_read = ?", true)
	case "unread":
		query = query.Where("is_read = ?", false)
	}

	limit := 100
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 || parsedLimit > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit notifikasi harus di antara 1 sampai 200"})
			return
		}
		limit = parsedLimit
	}

	var notifications []models.Notification
	if err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil notifikasi supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   notifications,
	})
}

func MarkSupplierNotificationRead(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	notificationID := c.Param("id")
	now := time.Now()

	result := config.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ? AND role = ?", notificationID, supplierID, string(models.RoleSupplier)).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menandai notifikasi"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notifikasi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notifikasi sudah dibaca",
	})
}

func GetSupplierOrders(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var orders []models.Order
	query := config.DB.
		Preload("Product").
		Preload("Umkm").
		Where("supplier_id = ?", supplierID)

	if status := strings.TrimSpace(c.Query("status")); status != "" && status != "all" {
		normalizedStatus, valid := (&models.Order{}).NormalizeStatus(status)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status pesanan tidak valid"})
			return
		}
		query = query.Where("status = ?", normalizedStatus)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar pesanan supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   orders,
	})
}

func UpdateOrderStatus(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var input updateSupplierOrderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID dan status wajib diisi"})
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		orderID = input.OrderID
	}
	if strings.TrimSpace(orderID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID wajib diisi"})
		return
	}

	nextStatus, valid := (&models.Order{}).NormalizeStatus(input.Status)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status pesanan tidak valid"})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Product").Where("id = ? AND supplier_id = ?", orderID, supplierID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pesanan"})
		return
	}

	previousStatus := order.Status
	if err := order.TransitionTo(nextStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          err.Error(),
			"current_status": previousStatus,
			"next_status":    nextStatus,
		})
		return
	}

	if err := config.DB.Model(&order).Update("status", order.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status pesanan"})
		return
	}

	shortOrderID := order.ID
	if len(shortOrderID) > 8 {
		shortOrderID = shortOrderID[:8]
	}
	orderLabel := "ORD-" + strings.ToUpper(shortOrderID)
	description := "Status pesanan " + orderLabel + " untuk produk " + order.Product.Name + " diubah dari " + string(previousStatus) + " menjadi " + string(nextStatus)
	_ = services.CreateActivityLog(nil, supplierID, "UPDATE_ORDER_STATUS", description)
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     supplierID,
		Role:       string(models.RoleSupplier),
		Title:      "Status Pesanan Diperbarui",
		Message:    description,
		Type:       "order_status_updated",
		SourceType: "order",
		SourceID:   order.ID,
	})
	_, _ = services.CreateActivityNotification(nil, models.Notification{
		UserID:     order.UmkmID,
		Role:       string(models.RoleUser),
		Title:      "Status Pesanan Diperbarui",
		Message:    description,
		Type:       "order_status_updated",
		SourceType: "order",
		SourceID:   order.ID,
	})
	_, _ = services.CreateRoleActivityNotifications(nil, models.RoleAdmin, models.Notification{
		Title:      "Status Pesanan Supplier Diperbarui",
		Message:    description,
		Type:       "order_status_updated",
		SourceType: "order",
		SourceID:   order.ID,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status pesanan berhasil diperbarui",
		"data":    order,
	})
}

type PublicSupplierResponse struct {
	ID            string  `json:"id"`
	BusinessName  string  `json:"business_name"`
	Category      string  `json:"category"`
	Region        string  `json:"region"`
	RatingAverage float64 `json:"rating_average"`
	ReviewCount   int64   `json:"review_count"`
	ProductCount  int64   `json:"product_count"`
}

type SupplierDetailResponse struct {
	ID            string  `json:"id"`
	BusinessName  string  `json:"business_name"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	Category      string  `json:"category"`
	Region        string  `json:"region"`
	PICName       string  `json:"pic_name"`
	Phone         string  `json:"phone"`
	Status        string  `json:"status"`
	RatingAverage float64 `json:"rating_average"`
	ReviewCount   int64   `json:"review_count"`
	ProductCount  int64   `json:"product_count"`
}

// GetPublicSuppliers mengambil semua supplier terverifikasi/aktif
func GetPublicSuppliers(c *gin.Context) {
	var suppliers []models.User
	if err := config.DB.Where("role = ? AND status = ?", models.RoleSupplier, "active").Order("created_at DESC").Find(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar supplier"})
		return
	}

	response := make([]PublicSupplierResponse, 0, len(suppliers))
	for _, supplier := range suppliers {
		var productCount int64
		config.DB.Model(&models.Product{}).Where("supplier_id = ?", supplier.ID).Count(&productCount)

		var stats struct {
			Average float64
			Count   int64
		}
		config.DB.Model(&models.Review{}).
			Joins("JOIN products ON reviews.product_id = products.id").
			Where("products.supplier_id = ?", supplier.ID).
			Select("COALESCE(AVG(reviews.rating), 0) as average, COUNT(reviews.id) as count").
			Scan(&stats)

		response = append(response, PublicSupplierResponse{
			ID:            supplier.ID,
			BusinessName:  supplier.BusinessName,
			Category:      supplier.Category,
			Region:        supplier.Region,
			RatingAverage: stats.Average,
			ReviewCount:   stats.Count,
			ProductCount:  productCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetSupplierDetail mengambil data detail lengkap supplier (wajib login)
func GetSupplierDetail(c *gin.Context) {
	supplierID := c.Param("id")
	var supplier models.User
	if err := config.DB.Where("id = ? AND role = ?", supplierID, models.RoleSupplier).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data supplier"})
		return
	}

	var productCount int64
	config.DB.Model(&models.Product{}).Where("supplier_id = ?", supplier.ID).Count(&productCount)

	var stats struct {
		Average float64
		Count   int64
	}
	config.DB.Model(&models.Review{}).
		Joins("JOIN products ON reviews.product_id = products.id").
		Where("products.supplier_id = ?", supplier.ID).
		Select("COALESCE(AVG(reviews.rating), 0) as average, COUNT(reviews.id) as count").
		Scan(&stats)

	response := SupplierDetailResponse{
		ID:            supplier.ID,
		BusinessName:  supplier.BusinessName,
		Email:         supplier.Email,
		Address:       supplier.Address,
		Category:      supplier.Category,
		Region:        supplier.Region,
		PICName:       supplier.PICName,
		Phone:         supplier.Phone,
		Status:        supplier.Status,
		RatingAverage: stats.Average,
		ReviewCount:   stats.Count,
		ProductCount:  productCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

type supplierStockConfirmationInput struct {
	Action string `json:"action" binding:"required"`
	Note   string `json:"note"`
}

type supplierHubHTTPError struct {
	Status  int
	Message string
}

func (e supplierHubHTTPError) Error() string {
	return e.Message
}

func ConfirmSupplierHubStock(c *gin.Context) {
	supplierID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	orderID := strings.TrimSpace(c.Param("order_id"))
	var input supplierStockConfirmationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action confirm atau reject wajib diisi"})
		return
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	note := strings.TrimSpace(input.Note)
	if action != "confirm" && action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action harus bernilai confirm atau reject"})
		return
	}

	responseStatus := http.StatusOK
	responseMessage := ""
	var responseOrder models.Order
	var responsePayment *models.Payment
	var gatewayPayload interface{}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", orderID).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return supplierHubHTTPError{Status: http.StatusNotFound, Message: "Order tidak ditemukan"}
			}
			return err
		}

		if order.SupplierID != supplierID {
			return supplierHubHTTPError{Status: http.StatusForbidden, Message: "Supplier tidak boleh mengonfirmasi order supplier lain"}
		}

		if action == "reject" {
			order.Status = models.OrderRejectedBySupplier
			if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
				return err
			}
			if err := services.CreateFinanceLog(tx, order, nil, "supplier_rejected", note); err != nil {
				return err
			}
			responseOrder = order
			responseMessage = "Order ditolak supplier"
			return nil
		}

		if order.StockDeducted {
			return supplierHubHTTPError{Status: http.StatusBadRequest, Message: "Stok untuk order ini sudah pernah dikurangi."}
		}

		var product models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ProductID).
			First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return supplierHubHTTPError{Status: http.StatusNotFound, Message: "Produk tidak ditemukan"}
			}
			return err
		}

		if product.Stock < order.Quantity {
			order.Status = models.OrderStockUnavailable
			if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
				return err
			}
			if err := services.CreateFinanceLog(tx, order, nil, "stock_not_enough", "Stok tidak mencukupi. "+note); err != nil {
				return err
			}
			responseStatus = http.StatusBadRequest
			responseMessage = "Stok tidak mencukupi."
			responseOrder = order
			return nil
		}

		product.Stock -= order.Quantity
		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		order.StockDeducted = true
		order.Status = models.OrderSupplierConfirmed
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"stock_deducted": true,
			"status":         order.Status,
		}).Error; err != nil {
			return err
		}
		if err := services.CreateFinanceLog(tx, order, nil, "supplier_confirmed", note); err != nil {
			return err
		}

		payment := models.Payment{
			OrderID:         order.ID,
			UserID:          order.UmkmID,
			Amount:          order.GrandTotal,
			SupplierFee:     order.SystemFee,
			Status:          models.PaymentPending,
			PaymentMethod:   "smartbank",
			ExternalOrderID: order.ID,
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		gatewayResult, gatewayErr := services.CreateSmartBankPaymentRequest(order, payment)
		gatewayPayload = gatewayResult
		if gatewayErr != nil {
			responseStatus = http.StatusBadGateway
			responseMessage = "Stok berhasil dikonfirmasi, tetapi payment request gagal dibuat"
			responseText := gatewayResult.RawResponse
			if responseText == "" {
				responseText = gatewayErr.Error()
			}

			payment.Status = models.PaymentFailed
			payment.GatewayStatus = "request_failed"
			payment.GatewayResponse = responseText
			if err := tx.Model(&payment).Updates(map[string]interface{}{
				"status":           payment.Status,
				"gateway_status":   payment.GatewayStatus,
				"gateway_response": payment.GatewayResponse,
			}).Error; err != nil {
				return err
			}

			order.Status = models.OrderPaymentRequestFailed
			if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
				return err
			}
			if err := services.CreateFinanceLog(tx, order, &payment, "payment_request_failed", gatewayErr.Error()); err != nil {
				return err
			}
			responseOrder = order
			responsePayment = &payment
			return nil
		}

		payment.PaymentReference = gatewayResult.PaymentReference
		payment.VirtualAccount = gatewayResult.VirtualAccount
		payment.GatewayStatus = gatewayResult.Status
		payment.GatewayResponse = gatewayResult.RawResponse
		if err := tx.Model(&payment).Updates(map[string]interface{}{
			"payment_reference": payment.PaymentReference,
			"virtual_account":   payment.VirtualAccount,
			"gateway_status":    payment.GatewayStatus,
			"gateway_response":  payment.GatewayResponse,
		}).Error; err != nil {
			return err
		}

		order.Status = models.OrderPaymentPending
		if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
			return err
		}
		if err := services.CreateFinanceLog(tx, order, &payment, "payment_request_created", "Payment request berhasil dibuat melalui API Gateway"); err != nil {
			return err
		}

		responseMessage = "Stok dikonfirmasi dan payment request berhasil dibuat"
		responseOrder = order
		responsePayment = &payment
		return nil
	})

	if err != nil {
		var httpErr supplierHubHTTPError
		if errors.As(err, &httpErr) {
			c.JSON(httpErr.Status, gin.H{"error": httpErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses konfirmasi stok"})
		return
	}

	payload := gin.H{
		"status":  "success",
		"message": responseMessage,
		"data": gin.H{
			"order": responseOrder,
		},
	}
	if responsePayment != nil {
		payload["data"].(gin.H)["payment"] = responsePayment
	}
	if gatewayPayload != nil {
		payload["data"].(gin.H)["gateway_response"] = gatewayPayload
	}
	if responseStatus >= http.StatusBadRequest {
		payload["status"] = "warning"
	}

	c.JSON(responseStatus, payload)
}
