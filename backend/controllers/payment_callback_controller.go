package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"supplierhub-backend/dto"
	"supplierhub-backend/services"

	"github.com/gin-gonic/gin"
)

// HandleSupplierHubPaymentCallback handles callback request from Payment Service
func HandleSupplierHubPaymentCallback(c *gin.Context) {
	if !validateSupplierHubCallbackKey(c) {
		return
	}

	var payload dto.PaymentCallbackPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload callback pembayaran tidak valid: " + err.Error()})
		return
	}

	// Delegasikan SEMUA proses ke PaymentService
	payment, order, shipmentLog, shipmentWarning, err := services.NewPaymentService().ProcessCallback(payload)
	if err != nil {
		var serviceErr services.PaymentServiceError
		if errors.As(err, &serviceErr) {
			c.JSON(serviceErr.Status, gin.H{"error": serviceErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses callback pembayaran: " + err.Error()})
		return
	}

	response := gin.H{
		"status":  "success",
		"message": "Callback pembayaran berhasil diproses",
		"data": gin.H{
			"payment": payment,
			"order":   order,
		},
	}
	if shipmentLog != nil {
		response["data"].(gin.H)["shipment"] = shipmentLog
	}
	if shipmentWarning != "" {
		response["status"] = "warning"
		response["message"] = "Pembayaran sukses, tetapi integrasi LogistiKita gagal"
		response["data"].(gin.H)["shipment_error"] = shipmentWarning
	}

	c.JSON(http.StatusOK, response)
}

func validateSupplierHubCallbackKey(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("SUPPLIERHUB_CALLBACK_API_KEY"))
	if expected == "" {
		return true
	}

	provided := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
	if provided == "" {
		provided = strings.TrimSpace(c.GetHeader("X-Callback-Token"))
	}
	if provided != expected {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Callback token tidak valid"})
		return false
	}
	return true
}
