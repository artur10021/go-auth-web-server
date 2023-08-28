package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	//hashOfRefreshToken, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	coll := client.Database("webGoServer").Collection("refreshTokens")
	var hashOfRefreshTokenFromDB bson.M
	errFromDB := coll.FindOne(context.TODO(), bson.D{{Key: "guid", Value: guidParam}}).Decode(&hashOfRefreshTokenFromDB)
	if errFromDB != nil {
		log.Println("Refresh tokens did not match")
		log.Println(errFromDB)
		log.Println(hashOfRefreshTokenFromDB)
		return false
	} else {
		log.Println("hashOfRefreshTokenFromDB")
		log.Println(hashOfRefreshTokenFromDB["refreshToken"])

		err := bcrypt.CompareHashAndPassword([]byte(hashOfRefreshTokenFromDB["refreshToken"].(string)), []byte(refreshToken))
		if err != nil {
			log.Println("Refresh tokens does not match")
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
	//oldHashOfRefreshTokenInDB, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	newRefreshToken := createRefreshToken()
	newHashOfRefreshToken, _ := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
	coll := client.Database("webGoServer").Collection("refreshTokens")
	//var hashOfRefreshTokenFromDB bson.M
	result, err := coll.UpdateOne(
		context.TODO(),
		bson.D{{Key: "guid", Value: guidParam}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "refreshToken", Value: string(newHashOfRefreshToken)}}}},
	)
	fmt.Println(result)
	fmt.Println(string(newHashOfRefreshToken))

	fmt.Println("result")
	return newRefreshToken, err
}

func refreshTokensFunk() {
	http.HandleFunc("/refreshTokens", func(w http.ResponseWriter, r *http.Request) {
		// AccessAndRefreshMap := make(map,)
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
				//updateRefreshTokenInDB(guidParam, refreshToken)
			}
		} else {
			io.WriteString(w, "Refresh tokens or guid parametr does not match.")
		}

	})

}

func createNewAccessToken(w http.ResponseWriter, r *http.Request) string {
	var (
		key         []byte
		token       *jwt.Token
		stringToken string
		err         error
	)

	guidParam := r.URL.Query().Get("guid")
	key = []byte("secret-key")
	token = jwt.NewWithClaims(jwt.SigningMethodHS512,
		jwt.MapClaims{
			"timestamp": time.Now().Unix(),
			"guid":      guidParam,
		})

	// refreshToken := createRefreshToken()

	// refreshToken = jwt.NewWithClaims(jwt.SigningMethodHS512,
	// 	jwt.MapClaims{
	// 		"exp":  time.Now().Add(time.Hour * 24),
	// 		"guid": guidParam,
	// 	})

	stringToken, err = token.SignedString(key)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return err.Error()
	}
	return stringToken
}

func saveRefreshTokenToDB(guidParam string, refreshToken string) {
	///////////// func conectDB
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
		errors.New("Duplicate! This guid parameter is already registered")
	}

	// if err == mongo.ErrNoDocuments {
	// 	fmt.Printf("No document was found with the title %s\n", title)
	// 	return
	// }
	// if err != nil {
	// 	panic(err)
	// }
	// jsonData, err := json.MarshalIndent(result, "", "    ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%s\n", jsonData)

	/////////////////
	fmt.Println(guidParam)
	fmt.Println(string(hashOfRefreshToken))
}

func setTokens() {
	http.HandleFunc("/setToken", func(w http.ResponseWriter, r *http.Request) {
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

	})
}

func main() {

	setTokens()
	refreshTokensFunk()

	fmt.Println("server started on port 80")
	http.ListenAndServe(":80", nil)
}
