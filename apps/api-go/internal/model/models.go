package model

import "time"

type DeliveryRuleStatus string

const (
	DeliveryRuleStatusDraft    DeliveryRuleStatus = "draft"
	DeliveryRuleStatusActive   DeliveryRuleStatus = "active"
	DeliveryRuleStatusDisabled DeliveryRuleStatus = "disabled"
)

type RuleConditionType string

const (
	RuleConditionCountry     RuleConditionType = "country"
	RuleConditionPostalCode  RuleConditionType = "postal_code"
	RuleConditionCartTotal   RuleConditionType = "cart_total"
	RuleConditionProductTag  RuleConditionType = "product_tag"
	RuleConditionCustomerTag RuleConditionType = "customer_tag"
)

type RuleActionType string

const (
	RuleActionHide   RuleActionType = "hide"
	RuleActionRename RuleActionType = "rename"
	RuleActionSort   RuleActionType = "sort"
)

type Shop struct {
	ID        uint      `gorm:"primaryKey"`
	Domain    string    `gorm:"size:255;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	DeliveryRules   []DeliveryRule           `gorm:"foreignKey:ShopID"`
	AuditLogs       []AuditLog               `gorm:"foreignKey:ShopID"`
	ConfigSnapshots []FunctionConfigSnapshot `gorm:"foreignKey:ShopID"`
}

type DeliveryRule struct {
	ID             uint               `gorm:"primaryKey"`
	ShopID         uint               `gorm:"not null;index"`
	Shop           Shop               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name           string             `gorm:"size:255;not null"`
	Priority       int                `gorm:"not null;default:100;index"`
	Status         DeliveryRuleStatus `gorm:"size:32;not null;default:draft;index"`
	ConditionType  RuleConditionType  `gorm:"size:64;not null"`
	ConditionValue string             `gorm:"type:text;not null"`
	ActionType     RuleActionType     `gorm:"size:64;not null"`
	ActionValue    string             `gorm:"type:text;not null"`
	CreatedAt      time.Time          `gorm:"not null"`
	UpdatedAt      time.Time          `gorm:"not null"`

	AuditLogs []AuditLog `gorm:"foreignKey:DeliveryRuleID"`
}

type AuditLog struct {
	ID             uint  `gorm:"primaryKey"`
	ShopID         uint  `gorm:"not null;index"`
	Shop           Shop  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DeliveryRuleID *uint `gorm:"index"`
	DeliveryRule   *DeliveryRule
	Event          string    `gorm:"size:128;not null;index"`
	Message        string    `gorm:"type:text;not null"`
	CreatedAt      time.Time `gorm:"not null;index"`
}

type FunctionConfigSnapshot struct {
	ID         uint      `gorm:"primaryKey"`
	ShopID     uint      `gorm:"not null;index"`
	Shop       Shop      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ConfigJSON string    `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time `gorm:"not null;index"`
}
