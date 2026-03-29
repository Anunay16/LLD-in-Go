package main

import (
	"fmt"
	"singleton_pattern/database"
)

func main() {
	dbInstance := database.GetSingletonDatabaseInstance()
	name := dbInstance.GetUserName(1)
	fmt.Print(name)
}
