package main

import (
	"flag"
	"fmt"
	migration "url-shortener/m/internal"
)

func main() {

	upFlag := flag.Bool("up", false, "Run migrations UP")
	downFlag := flag.Bool("down", false, "Run migrations DOWN")
	flag.Parse()

	if *upFlag {
		fmt.Println("Running Migrations UP...")
		if err := migration.Up(); err != nil {
			fmt.Printf("Error running migrations UP: %v\n", err)
			return
		}
		fmt.Println("Migrations executed!")
		return
	}

	if *downFlag {
		fmt.Println("Running Migrations DOWN...")
		if err := migration.Down(); err != nil {
			fmt.Printf("Error running migrations DOWN: %v\n", err)
			return
		}
		fmt.Println("Migrations executed!")
		return
	}

	var number string
	fmt.Println("Migrations CLI")
	fmt.Println("Type the number of the command desired:\n1-Migrations UP\n2-Migrations DOWN\n3-Create a new migration")
	_, err := fmt.Scan(&number)
	if err != nil {
		fmt.Println("Error while reading the values", err)
	}

	if number == "1" {
		fmt.Println("Running Migrations UP...")
		if err := migration.Up(); err != nil {
			fmt.Printf("Error running migrations UP: %v\n", err)
			return
		}
		fmt.Println("Migrations executed!")
		return
	}

	if number == "2" {
		fmt.Println("Running Migrations DOWN...")
		if err := migration.Down(); err != nil {
			fmt.Printf("Error running migrations DOWN: %v\n", err)
			return
		}
		fmt.Println("Migrations executed!")
		return
	}

	if number == "3" {
		fmt.Println("Type the name of the migration desired:")
		var name string
		_, err := fmt.Scan(&name)
		if err != nil {
			fmt.Println("Error while reading the values", err)
		}
		fmt.Println("Creating a new migration...")
		if err := migration.Create(name); err != nil {
			fmt.Printf("Error creating migration: %v\n", err)
			return
		}
		fmt.Println("Migration created!")
	}

}
