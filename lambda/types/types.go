package types

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type SecretManClient struct {
	secretStore *secretsmanager.Client
}

func NewSecretClient() SecretManClient {
	region := "ap-south-1"
    
	// need to do like this as module are pre-executed before lambda evn start so, causes the timeout 
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		fmt.Println(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	return SecretManClient{
		secretStore : svc,
	}

}


func (s SecretManClient) GetTheSecret() string {
	secretName := "JWTSecret"

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	result,err:= s.secretStore.GetSecretValue(context.TODO(),input)
	if err != nil {
		fmt.Println(err)
	}

	// Print the secret value
	secretValue := *result.SecretString
	fmt.Println("Secret Value:", secretValue)

	return secretValue
}


type RegisterUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}

func NewUser(registerUser RegisterUser) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUser.Password), 11)

	if err != nil {
		return User{}, err
	}

	return User{
		Username:     registerUser.Username,
		PasswordHash: string(hashedPassword),
	}, nil

}

func ValidatePassword(hashedPassword, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword))

	return err == nil
}

func CreateToken(user User) string {

	now := time.Now()
	validUntil := now.Add(time.Hour * 1).Unix()

	claims := jwt.MapClaims{
		"user":    user.Username,
		"expires": validUntil,
	}
 

	sClient := NewSecretClient()
	secret:=sClient.GetTheSecret()
	fmt.Printf("x: %v, type: %T\n", secret, secret) // this is just for debug


	// success 

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims, nil)
	

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	return tokenString
}
