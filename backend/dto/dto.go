package dto

import "mime/multipart"

// RegisterInput represents registration inputs from Form or JSON
type RegisterInput struct {
	BusinessName string                `form:"business_name" json:"business_name" binding:"required"`
	Email        string                `form:"email" json:"email" binding:"required,email"`
	Password     string                `form:"password" json:"password"`
	Role         string                `form:"role" json:"role" binding:"required"`
	Address      string                `form:"address" json:"address"`
	Category     string                `form:"category" json:"category"`
	Region       string                `form:"region" json:"region"`
	IsGoogle     bool                  `form:"is_google" json:"is_google"`
	Document     *multipart.FileHeader `form:"document" json:"document"`
}

// LoginInput represents login input credentials
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// PaymentCallbackPayload represents callback payload from Payment Gateways (e.g. SmartBank)
type PaymentCallbackPayload struct {
	PaymentReference string `json:"payment_reference"`
	ExternalOrderID  string `json:"external_order_id"`
	OrderID          string `json:"order_id"`
	PaymentStatus    string `json:"payment_status"`
	Status           string `json:"status"`
	PaidAt           string `json:"paid_at"`
}
