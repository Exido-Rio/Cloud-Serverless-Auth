package api

import (
	"encoding/json"
	"fmt"
	"lambda-func/database"
	"lambda-func/types"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ApiHandler struct {
	dbStore database.UserStore
}

func NewApiHandler(dbStore database.UserStore) ApiHandler {

	return ApiHandler{
		dbStore: dbStore,
	}

}

func (api ApiHandler) RegisterUserHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var registerUser types.RegisterUser

	err := json.Unmarshal([]byte(request.Body), &registerUser)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid Request ",
			StatusCode: http.StatusBadRequest,
		}, err
	}

	if registerUser.Username == "" || registerUser.Password == "" {

		return events.APIGatewayProxyResponse{
			Body:       "Field Empty",
			StatusCode: http.StatusBadRequest,
		}, err

	}

	userExist, err := api.dbStore.DoesUserExist(registerUser.Username)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Internal Server Error ",
			StatusCode: http.StatusInternalServerError,
		}, err

	}

	if userExist {
		return events.APIGatewayProxyResponse{
			Body:       "User Already Exist",
			StatusCode: http.StatusConflict,
		}, nil
	}

	user, err := types.NewUser(registerUser)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Internal Server Error ",
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("	Could not create user - %w", err)
	}

	err = api.dbStore.InsertUser(user)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Internal Server Error ",
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("error inserting user - %w", err)

	}

	return events.APIGatewayProxyResponse{
		Body:       "User Registered SuccessFully ",
		StatusCode: http.StatusOK,
	}, nil
}

func (api ApiHandler) LoginUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var loginRequest LoginRequest


	err := json.Unmarshal([]byte(request.Body),&loginRequest)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Invalid Request ",
			StatusCode: http.StatusBadRequest,
		}, err
	}


	user , err := api.dbStore.GetUser(loginRequest.Username)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "Internal Server Error ",
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	if !types.ValidatePassword(user.PasswordHash,loginRequest.Password) { return events.APIGatewayProxyResponse{
			Body:       "Invalid User Credential ",
			StatusCode: http.StatusBadRequest,
		}, err
	}


	accessToken := types.CreateToken(user) 
	
	successMsg := fmt.Sprintf(`{access_token : %s}`,accessToken)

    return events.APIGatewayProxyResponse{
		Body:     successMsg ,
		StatusCode: http.StatusOK,
	}, err
}
