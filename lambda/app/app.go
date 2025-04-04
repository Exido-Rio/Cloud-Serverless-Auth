package app

import (
	"lambda-func/api"
	"lambda-func/database"
)

type App struct {
	ApiHandler api.ApiHandler
}


func NewApp() App {
	// init our dbStore 
	// get passed down into api handler 
	db := database.NewDynamoDbClient()
	apiHandler := api.NewApiHandler(db)

	return App{
		ApiHandler: apiHandler,
	}
}