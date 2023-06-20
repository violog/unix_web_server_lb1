package main

import (
	"context"
	"os"

	"go.uber.org/zap"

	"todo"
	"todo/database"
)

func main() {
	log := zap.NewExample()
	ctx := context.Background()
	db, err := database.New("postgres://postgres:123456@db:5432/todo?sslmode=disable")
	if err != nil {
		log.Error("could not create database" + err.Error())
		os.Exit(1)
	}

	app, err := todo.New(db)
	if err != nil {
		log.Error("could not create representation of app" + err.Error())
		os.Exit(1)
	}

	err = app.Run(ctx)
	if err != nil {
		log.Error("could not run app" + err.Error())
		os.Exit(1)
	}

	//err = app.Close()
	//if err != nil {
	//	log.Error("could not close app" + err.Error())
	//	os.Exit(1)
	//}
}
