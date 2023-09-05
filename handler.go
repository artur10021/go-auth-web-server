package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setTokens(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	coll := client.Database("webGoServer").Collection("refreshTokens")
	var result bson.M
	guidParam := r.URL.Query().Get("guid")
	coll.FindOne(context.TODO(), bson.D{{Key: "guid", Value: guidParam}}).Decode(&result)
	if len(result) == 0 {
		accessToken := string(createNewAccessToken(w, r))
		refreshToken := createRefreshToken()
		jsonAccessAndRefreshTokens, errorFromJson := json.Marshal(map[string]string{"accessToken": accessToken, "refreshToken": refreshToken})
		if errorFromJson != nil {
			http.Error(w, "Converting json error", http.StatusInternalServerError)
			return
		}

		saveRefreshTokenToDB(guidParam, refreshToken)
		w.Write(jsonAccessAndRefreshTokens)
	} else {
		log.Println("Duplicate! This guid parameter is already registered")
		io.WriteString(w, "Duplicate! This guid parameter is already registered")
	}
}

func refreshTokensFunk(w http.ResponseWriter, r *http.Request) {

	oldRefreshToken := r.URL.Query().Get("refreshToken")
	guidParam := r.URL.Query().Get("guid")
	isValidateRefreshToken := isValidateTokenFunk(guidParam, oldRefreshToken)
	if isValidateRefreshToken {
		newRefreshToken, err := updateRefreshTokenInDB(guidParam, oldRefreshToken)
		if err == nil {
			accessToken := string(createNewAccessToken(w, r))
			refreshToken := newRefreshToken
			jsonAccessAndRefreshTokens, errorFromJson := json.Marshal(map[string]string{"accessToken": accessToken, "refreshToken": refreshToken})
			if errorFromJson != nil {
				http.Error(w, "Converting json error", http.StatusInternalServerError)
				return
			}
			w.Write((jsonAccessAndRefreshTokens))
		}
	} else {
		io.WriteString(w, "Refresh tokens or guid parametr does not match.")
	}
}
