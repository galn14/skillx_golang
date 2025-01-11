package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transactionInput struct {
		UserId    string `json:"user_id"`
		ProductId string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	// Decode JSON input
	if err := json.NewDecoder(r.Body).Decode(&transactionInput); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validate input
	if transactionInput.UserId == "" || transactionInput.ProductId == "" || transactionInput.Quantity <= 0 {
		utils.RespondError(w, http.StatusBadRequest, "User ID, Product ID, and quantity are required")
		return
	}

	// Initialize Firebase client
	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch product details
	productRef := client.NewRef("products/" + transactionInput.UserId + "/" + transactionInput.ProductId)
	var product models.Product
	if err := productRef.Get(ctx, &product); err != nil {
		fmt.Printf("Error fetching product from Firebase: %v\n", err)
		utils.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Log product data for debugging
	fmt.Printf("Fetched product: %+v\n", product)

	// Validate product data
	if product.Price == "" {
		utils.RespondError(w, http.StatusBadRequest, "Product price is missing")
		return
	}

	// Normalize and parse price
	normalizedPrice := strings.ReplaceAll(product.Price, ".", "")
	normalizedPrice = strings.ReplaceAll(normalizedPrice, ",", "")
	pricePerUnit, err := strconv.ParseFloat(normalizedPrice, 64)
	if err != nil {
		fmt.Printf("Error parsing price: %v (Raw Price: %s)\n", err, product.Price)
		utils.RespondError(w, http.StatusInternalServerError, "Invalid product price format")
		return
	}

	// Calculate total price
	totalPrice := pricePerUnit * float64(transactionInput.Quantity)

	// Create a new transaction
	transaction := models.Transaction{
		IdTransaction:   uuid.New().String(),
		UserId:          transactionInput.UserId,
		SellerId:        transactionInput.UserId,
		ProductId:       transactionInput.ProductId,
		Price:           fmt.Sprintf("%.2f", pricePerUnit),
		TotalPrice:      fmt.Sprintf("%.2f", totalPrice),
		Quantity:        transactionInput.Quantity,
		Status:          "pending",
		TransactionTime: time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Midtrans Snap API integration
	if config.GlobalMidtransConfig == nil || config.GlobalMidtransConfig.SnapClient == nil {
		fmt.Println("Midtrans SnapClient is not initialized")
		utils.RespondError(w, http.StatusInternalServerError, "Payment gateway not configured")
		return
	}

	snapClient := config.GlobalMidtransConfig.SnapClient

	// Prepare Snap request
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transaction.IdTransaction,
			GrossAmt: int64(totalPrice),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: "customer@example.com",
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    product.UID,
				Price: int64(pricePerUnit),
				Qty:   int32(transactionInput.Quantity),
				Name:  product.NameProduct,
			},
		},
	}

	// Log Snap request for debugging
	fmt.Printf("Snap Request: %+v\n", snapReq)

	// Create transaction on Midtrans
	snapResp, err := snapClient.CreateTransaction(snapReq)
	if err != nil {
		fmt.Printf("Error creating Snap transaction: %v\n", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create transaction with payment gateway")
		return
	}

	// Update transaction with Midtrans details
	transaction.PaymentToken = snapResp.Token
	transaction.PaymentUrl = snapResp.RedirectURL

	// Save transaction to Firebase
	transactionRef := client.NewRef("transactions/" + transaction.UserId + "/" + transaction.IdTransaction)
	if err := transactionRef.Set(ctx, &transaction); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to store transaction data")
		return
	}

	// Log successful transaction creation
	fmt.Println("Transaction created successfully")

	// Respond with Snap token and redirect URL
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"transaction": transaction,
			"url":         snapResp.RedirectURL,
		},
		"message": "Transaction created successfully",
	})
}
