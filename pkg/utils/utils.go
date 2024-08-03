package utils

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func ConnectToDB() (*sql.DB, error) {
	err := godotenv.Load("../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
		return nil, err
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDb := os.Getenv("POSTGRES_DATABASE")
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")

	//build connection string
	postgresUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDb)

	db, err := sql.Open("postgres", postgresUrl)

	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to database", postgresUrl)
	return db, err
}
