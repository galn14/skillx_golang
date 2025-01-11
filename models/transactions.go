package models

import "time"

type Transaction struct {
	IdTransaction   string    `json:"id_transaction" gorm:"primaryKey"`
	UserId          string    `json:"user_id"`                   // ID Pembeli
	SellerId        string    `json:"seller_id"`                 // ID Penjual
	ProductId       string    `json:"product_id"`                // ID Produk yang dibeli
	Price           string    `json:"price"`                     // Harga satuan produk
	TotalPrice      string    `json:"total_price"`               // Total harga transaksi
	Quantity        int       `json:"quantity"`                  // Jumlah produk
	Status          string    `json:"status"`                    // Status transaksi: pending, paid, failed, cancelled
	PaymentType     string    `json:"payment_type"`              // Jenis pembayaran (e.g., gopay, credit_card)
	TransactionTime time.Time `json:"transaction_time"`          // Waktu transaksi
	SettlementTime  time.Time `json:"settlement_time,omitempty"` // Waktu penyelesaian pembayaran
	OrderId         string    `json:"order_id"`                  // Order ID dari Midtrans
	PaymentToken    string    `json:"payment_token"`             // Token pembayaran Midtrans
	PaymentUrl      string    `json:"payment_url"`               // URL pembayaran Midtrans
	SnapResponse    string    `json:"snap_response"`             // Respons Snap API (disimpan untuk log/debug)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
