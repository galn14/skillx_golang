package controllers

import (
	"context"
	"golang-firebase-backend/config"
	"golang-firebase-backend/utils"
	"net/http"

	"google.golang.org/api/iterator"
)

func ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Storage(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Storage")
		return
	}

	bucketHandle, err := client.Bucket("skillx-butterscoth.firebasestorage.app")
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to get bucket handle")
		return
	}

	bucket := bucketHandle.Objects(ctx, nil)

	files := []string{}
	for {
		obj, err := bucket.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to iterate files")
			return
		}
		files = append(files, obj.Name)
	}

	utils.RespondJSON(w, http.StatusOK, files)
}
