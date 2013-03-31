package main

import "fmt"
import "mylibrary"

func main() {
	var message string
	message = "Hello world"
	fmt.Println(message)
	fmt.Println(mylibrary.PrintMessage("pete"))
}
