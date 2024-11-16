package controllers

import (
	"context"
	"golang-firebase-backend/config"
	"golang-firebase-backend/utils"
	"net/http"
)

func GetData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("users")
	var data map[string]interface{}
	if err := ref.Get(ctx, &data); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch data")
		return
	}

	utils.RespondJSON(w, http.StatusOK, data)
}
