package main

import (
	"fmt"
	"lambda-func/app"
	"lambda-func/middleware"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Username string `json:"username"`
	//Password string `json:"password"`
}

// take  in a payload and do something
func HandleRequest(event MyEvent) (string,error) {

	if event.Username == "" {
		return "",fmt.Errorf("username cannot be empty")
	}

	return fmt.Sprintf("Successfully called  by - %s",event.Username),nil

}

func ProtectedHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)  {
	return events.APIGatewayProxyResponse{
		Body:       "Secret Route Unlocked!",
		StatusCode: http.StatusOK,
	},nil
}

func main() {
	myApp := app.NewApp()
	lambda.Start(func (request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)  {
		switch request.Path {
		case "/register" :
			return myApp.ApiHandler.RegisterUserHandler(request)

		case "/login" :
			return myApp.ApiHandler.LoginUser(request)
		case "/protected" :
			return middleware.ValidateJWTMiddleware(ProtectedHandler)(request)
		default :
		    return events.APIGatewayProxyResponse{
			Body: "Not Found ",
			StatusCode: http.StatusNotFound,
		},nil
		}
		
	})

}