package types

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	AdminUser   = "admin"
	KitchenUser = "kitchen"
)

type userRole string
type User struct {
	Pk             string   `json:"-" dynamodbav:"pk"` // user
	Sk             string   `json:"id" dynamodbav:"sk"`
	HashedPassword []byte   `json:"hashedPassword" dynamodbav:"hashedPassword"` // <hashedPassword>
	Role           userRole `json:"role" dynamodbav:"role"`
}
type UserTokenType struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
type Product struct {
	Pk          string `json:"-" dynamodbav:"pk"`
	Sk          string `json:"id" dynamodbav:"sk"`
	Title       string `json:"title" dynamodbav:"title,omitempty"`
	Price       uint16 `json:"price" dynamodbav:"price, omitempty"`
	Description string `json:"description,omitempty" dynamodbav:"description, omitempty"`
	Image       string `json:"image" dynamodbav:"image, omitempty"`
}

type Item struct {
	Id    string `json:"id" dynamodbav:"id"`
	Title string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Price uint16 `json:"price" dynamodbav:"price, omitempty"`
	Qty   uint8  `json:"qty" dynamodbav:"qty, omitempty"`
}

type Order struct {
	Pk        string `json:"-" dynamodbav:"pk"`
	Sk        string `json:"id" dynamodbav:"sk"`
	Items     []Item `json:"items" dynamodbav:"items"`
	Notes     string `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
	SessionId string `json:"sessionId" dynamodbav:"sessionId"`
	CreatedAt int64  `json:"createdAt" dynamodbav:"createdAt"`
}

type Table struct {
	Pk               string `json:"-" dynamodbav:"pk"`
	Sk               string `json:"id" dynamodbav:"sk"`
	Title            string `json:"title" dynamodbav:"title"`
	IsAvailable      bool   `json:"isAvailable" dynamodbav:"isAvailable"`
	CurrentSessionId string `json:"currentSession,omitempty" dynamodbav:"currentSession,omitempty"`
	UpdatedAt        int64  `json:"updatedAt" dynamodbav:"updatedAt"`
}

type sessionStatus string

const (
	SessionOngoing  = "ongoing"
	SessionFinished = "finished"
)

type BackendSession struct {
	Pk        string   `json:"-" dynamodbav:"pk"` //session:backend
	Sk        string   `json:"id" dynamodbav:"sk"`
	CsrfToken string   `json:"csrfToken" dynamodbav:"csrfToken"`
	Role      userRole `json:"role" dynamodbav:"role"`
	Name      string   `json:"name" dynamodbav:"name"`
	CreatedAt int64    `json:"createdAt" dynamodbav:"createdAt"`
}
type Session struct {
	Pk        string        `json:"-" dynamodbav:"pk"` //session
	Sk        string        `json:"id" dynamodbav:"sk"`
	Name      string        `json:"name" dynamodbav:"name"`
	Role      userRole      `json:"role" dynamodbav:"role"`
	Phone     int64         `json:"phone" dynamodbav:"phone"`
	Status    sessionStatus `json:"status" dynamodbav:"status"`
	TableId   string        `json:"tableId" dynamodbav:"tableId"`
	Orders    []string      `json:"orders" dynamodbav:"orders"`
	Amount    uint16        `json:"amount,omitempty" dynamodbav:"amount,omitempty"`
	CsrfToken string        `json:"csrfToken" dynamodbav:"csrfToken"`
	CreatedAt int64         `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt int64         `json:"updatedAt" dynamodbav:"updatedAt"`
}

type TableToken struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}
