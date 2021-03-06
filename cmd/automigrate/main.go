package main

import (
	dbModule "github.com/notegio/openrelay/db"
	"github.com/notegio/openrelay/common"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
	"fmt"
	"strings"
)

func main() {
	pgHost := os.Args[1]
	pgUser := os.Args[2]
	pgPassword := common.GetSecret(os.Args[3])
	connectionString := fmt.Sprintf(
		"host=%v dbname=postgres sslmode=disable user=%v password=%v",
		pgHost,
		pgUser,
		pgPassword,
	)
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Could not open postgres connection: %v", err.Error())
	}
	if err := db.AutoMigrate(&dbModule.Order{}).Error; err != nil {
		log.Fatalf("Error migrating database: %v", err.Error())
	}
	for _, credString := range(os.Args[4:]) {
		creds := strings.Split(credString, ";")
		if len(creds) != 3 {
			log.Printf("Malformed credential string: %v", credString)
			continue
		}
		username, passwordURI, permissions := creds[0], creds[1], creds[2]
		password := common.GetSecret(passwordURI)
		// I don't like using string formatting instead of paramterization, but I
		// don't know of a way to parameterize the username in this statement. It
		// should still be fairly safe, because if you're able to execute this
		// command you already have administrative database access.
		if err = db.Exec(fmt.Sprintf("CREATE USER %v WITH PASSWORD '%v'", username, password)).Error; err != nil {
			log.Printf(err.Error())
		}
		for _, permission := range(strings.Split(permissions, ",")) {
			permArray := strings.Split(permission, ".")
			if len(permArray) != 2 {
				log.Printf("Malformed permission string '$v'", permission)
				continue
			}
			table, permission := permArray[0], permArray[1]
			// I don't like using string formatting instead of paramterization, but I
			// don't know of a way to parameterize the elements in this statement. It
			// should still be fairly safe, because if you're able to execute this
			// command you already have administrative database access.
			if err = db.Exec(fmt.Sprintf("GRANT %v ON TABLE %v TO %v", permission, table, username)).Error; err != nil {
				log.Printf(err.Error())
			}
		}
		log.Printf("Created '%v'", credString)
	}
}
