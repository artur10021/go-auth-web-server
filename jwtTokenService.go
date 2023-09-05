package main

import (
	"context"
	b64 "encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func createRefreshToken() string {
	uniqueToken := uuid.New().String()
	sEnc := b64.StdEncoding.EncodeToString([]byte(uniqueToken))
	return sEnc
}

func isValidateTokenFunk(guidParam string, refreshToken string) bool {
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
	var hashOfRefreshTokenFromDB bson.M
	errFromDB := coll.FindOne(context.TODO(), bson.D{{Key: "guid", Value: guidParam}}).Decode(&hashOfRefreshTokenFromDB)
	if errFromDB != nil {
		log.Println("Refresh tokens did not match")
		return false
	} else {
		err := bcrypt.CompareHashAndPassword([]byte(hashOfRefreshTokenFromDB["refreshToken"].(string)), []byte(refreshToken))
		if err != nil {
			log.Println("Refresh tokens did not match")
			return false
		}
	}
	return true
}

func updateRefreshTokenInDB(guidParam string, refreshToken string) (string, error) {
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
	newRefreshToken := createRefreshToken()
	newHashOfRefreshToken, _ := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
	coll := client.Database("webGoServer").Collection("refreshTokens")
	_, updateErr := coll.UpdateOne(
		context.TODO(),
		bson.D{{Key: "guid", Value: guidParam}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "refreshToken", Value: string(newHashOfRefreshToken)}}}},
	)
	return newRefreshToken, updateErr
}

func createNewAccessToken(w http.ResponseWriter, r *http.Request) string {
	var (
		key         []byte
		token       *jwt.Token
		stringToken string
		err         error
	)
	guidParam := r.URL.Query().Get("guid")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	key = []byte(os.Getenv("SECRET_KEY"))
	token = jwt.NewWithClaims(jwt.SigningMethodHS512,
		jwt.MapClaims{
			"timestamp": time.Now().Unix(),
			"guid":      guidParam,
		})
	stringToken, err = token.SignedString(key)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return err.Error()
	}
	return stringToken
}

func saveRefreshTokenToDB(guidParam string, refreshToken string) {
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
	coll.FindOne(context.TODO(), bson.D{{Key: "guid", Value: guidParam}}).Decode(&result)

	hashOfRefreshToken, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	if len(result) == 0 {
		coll.InsertOne(
			context.TODO(),
			bson.D{
				{Key: "guid", Value: guidParam},
				{Key: "refreshToken", Value: string(hashOfRefreshToken)},
			},
		)
		log.Println("The refresh token is saved in the DB.")
	} else {
		log.Println("Duplicate! This guid parameter is already registered")
	}
}
