package services

import (
	"errors"
	"strings"
	"time"

	"supplierhub-backend/config"
	"supplierhub-backend/dto"
	"supplierhub-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaymentServiceError represents service-level errors with HTTP status context
type PaymentServiceError struct {
	Status  int
	Message string
}

func (e PaymentServiceError) Error() string {
	return e.Message
}

// PaymentService handles payment-related business logic
type PaymentService struct {
	db *gorm.DB
}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService() *PaymentService {
	return &PaymentService{db: config.DB}
}

// ProcessCallback handles payment callback notifications from gateways
func (s *PaymentService) ProcessCallback(payload dto.PaymentCallbackPayload) (*models.Payment, *models.Order, *models.ShipmentLog, string, error) {
	callbackStatus := strings.ToLower(strings.TrimSpace(payload.Status))
	if callbackStatus == "" {
		callbackStatus = strings.ToLower(strings.TrimSpace(payload.PaymentStatus))
	}
	if callbackStatus == "" {
		return nil, nil, nil, "", PaymentServiceError{Status: 400, Message: "Status callback wajib diisi"}
	}

	isSuccess := callbackStatus == "success" || callbackStatus == "paid"
	isFailed := callbackStatus == "failed" || callbackStatus == "fail" || callbackStatus == "cancelled" || callbackStatus == "canceled"
	if !isSuccess && !isFailed {
		return nil, nil, nil, "", PaymentServiceError{Status: 400, Message: "Status callback harus success atau failed"}
	}

	var payment models.Payment
	var order models.Order
	var shipmentLog *models.ShipmentLog
	var shipmentWarning string

	callbackPayloadStr := JSONString(payload)
	paidAt := parseCallbackPaidAt(payload.PaidAt, isSuccess)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Order.Product").Preload("Order.Supplier").Preload("Order.Umkm")
		switch {
		case strings.TrimSpace(payload.PaymentReference) != "":
			query = query.Where("payment_reference = ?", strings.TrimSpace(payload.PaymentReference))
		case strings.TrimSpace(payload.ExternalOrderID) != "":
			query = query.Where("external_order_id = ? OR order_id = ?", strings.TrimSpace(payload.ExternalOrderID), strings.TrimSpace(payload.ExternalOrderID))
		case strings.TrimSpace(payload.OrderID) != "":
			query = query.Where("order_id = ?", strings.TrimSpace(payload.OrderID))
		default:
			return PaymentServiceError{Status: 400, Message: "payment_reference atau external_order_id wajib diisi"}
		}

		if err := query.First(&payment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return PaymentServiceError{Status: 404, Message: "Payment tidak ditemukan"}
			}
			return err
		}

		order = payment.Order
		if order.ID == "" {
			return PaymentServiceError{Status: 404, Message: "Order payment tidak ditemukan"}
		}

		if isFailed {
			payment.Status = models.PaymentFailed
			payment.CallbackStatus = "failed"
			payment.GatewayStatus = "failed"
			payment.GatewayResponse = callbackPayloadStr
			order.Status = models.OrderPaymentFailed

			if err := tx.Model(&payment).Updates(map[string]interface{}{
				"status":           payment.Status,
				"callback_status":  payment.CallbackStatus,
				"gateway_status":   payment.GatewayStatus,
				"gateway_response": payment.GatewayResponse,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
				return err
			}
			return CreateFinanceLog(tx, order, &payment, "payment_failed", "Callback SmartBank menyatakan pembayaran gagal")
		}

		payment.Status = models.PaymentSuccess
		payment.CallbackStatus = "success"
		payment.GatewayStatus = "success"
		payment.GatewayResponse = callbackPayloadStr
		payment.PaidAt = paidAt
		order.Status = models.OrderPaid

		if err := tx.Model(&payment).Updates(map[string]interface{}{
			"status":           payment.Status,
			"callback_status":  payment.CallbackStatus,
			"gateway_status":   payment.GatewayStatus,
			"gateway_response": payment.GatewayResponse,
			"paid_at":          payment.PaidAt,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
			return err
		}
		if err := CreateFinanceLog(tx, order, &payment, "payment_success", "Callback SmartBank sukses"); err != nil {
			return err
		}

		shipmentResult, shipmentErr := CreateLogistiKitaShipment(order)
		if shipmentErr != nil {
			shipmentWarning = shipmentErr.Error()
			if err := CreateShipmentLog(tx, order, shipmentResult, shipmentErr); err != nil {
				return err
			}
			return nil
		}

		order.Status = models.OrderShipmentCreated
		if err := tx.Model(&order).Update("status", order.Status).Error; err != nil {
			return err
		}
		if err := CreateFinanceLog(tx, order, &payment, "shipment_created", "Shipment berhasil dibuat di LogistiKita"); err != nil {
			return err
		}
		if err := CreateShipmentLog(tx, order, shipmentResult, nil); err != nil {
			return err
		}

		var savedShipment models.ShipmentLog
		if err := tx.Where("order_id = ?", order.ID).Order("created_at DESC").First(&savedShipment).Error; err == nil {
			shipmentLog = &savedShipment
		}
		return nil
	})

	if err != nil {
		return nil, nil, nil, "", err
	}

	return &payment, &order, shipmentLog, shipmentWarning, nil
}

func parseCallbackPaidAt(raw string, success bool) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if !success {
			return nil
		}
		now := time.Now()
		return &now
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return &parsed
		}
	}

	now := time.Now()
	return &now
}
