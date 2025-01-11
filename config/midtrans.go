package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransConfig struct {
	SnapClient *snap.Client
	ServerKey  string
	ClientKey  string
	BaseURL    string
}

var GlobalMidtransConfig *MidtransConfig

func LoadMidtransConfig() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	clientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	baseURL := os.Getenv("BASE_URL")

	if serverKey == "" || clientKey == "" || baseURL == "" {
		log.Fatal("Midtrans configuration missing: ensure MIDTRANS_SERVER_KEY, MIDTRANS_CLIENT_KEY, and BASE_URL are set")
	}

	// Initialize Snap Client
	snapClient := snap.Client{}
	snapClient.New(serverKey, midtrans.Sandbox)

	GlobalMidtransConfig = &MidtransConfig{
		SnapClient: &snapClient,
		ServerKey:  serverKey,
		ClientKey:  clientKey,
		BaseURL:    baseURL,
	}

	log.Println("Midtrans Snap Client initialized successfully")
}
