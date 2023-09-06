# go-auth-web-server
![images](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/e2ac0c84-6eaf-4063-b6cd-4f4ae877f430)

# Authentication Service

This project represents a part of an authentication service written in the Go programming language. The service provides two REST routes for token management: issuing a pair of Access and Refresh tokens, and performing a token refresh operation.

## **Tokens Requirements:**
Access token: JWT type, SHA512 algorithm. Not stored in the database.
Refresh token: Custom type, base64 format. Stored securely using bcrypt.

## Technologies

The following technologies are used in this project:

- **Go**
- **JWT (JSON Web Tokens)**
- **MongoDB**
## Getting Started
 **1. Clone the Repository:** >`git clone https://github.com/artur10021/go-auth-with-jwt-server.git`
 
 **2. Install Dependencies:** >`go mod download`

 **3. Configuration** > Create a `.env` file in the project directory and set the necessary environment variables: 
`MONGODB_URI=mongodb://localhost:27017`
`SECRET_KEY=SECRET_KEY_FOR_JWT`

![image](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/f7a9e306-39ed-48ec-a6e9-6f015ca6a4c6)

 **4. Run the Application** >`go run .`

## Routes

### 1. Issuing Access and Refresh Tokens

Route for getting a pair of Access and Refresh tokens for.

**URL:** `/getTokens?guid={GUID}`

**Method:** `POST`

**Query Parameters:**
- `guid` (required) - The user identifier (GUID) for which to obtain tokens.

### 2. Token Refresh

Route for performing a token refresh operation.

**URL:** `/refreshTokens?guid={GUID}&refreshToken={refreshToken}`

**Method:** `POST`

**Query Parameters:**
- `guid` - The user identifier (GUID) for which to obtain tokens.
- `refreshToken` - The token returned when using the getTokens route

After launching the project using the command `go run main.go`, we can utilize two endpoints. To perform testing, I am using Postman.
We use the URL `/getTokens?guid={GUID}` by passing the guid parameter, and then we receive a `JSON` response from the server containing the accessToken and refreshToken.
![image](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/2716f2d6-269a-4754-b501-40184c3150f7)

**MongoDB:**

![image](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/e29e6626-8246-4f11-a73c-45172eab2abb)

After obtaining the `accessToken` and `refreshToken`, we can make use of the second endpoint `/refreshTokens?guid={GUID}&refreshToken={refreshToken}` to refresh the refreshToken in the database and receive a new accessToken.
![image](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/7f559bc3-cce5-4f3f-8faf-c2c570f62b01)

**MongoDB:**

![image](https://github.com/artur10021/go-auth-with-jwt-server/assets/66840544/6fbc0f03-aebb-41c9-8c8d-3715e030ca06)

