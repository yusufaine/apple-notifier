package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	_ "github.com/yusufaine/apple-inventory-notifier/cloudfunction"
)

func main() {
	port := flag.String("port", "3000", "")
	flag.Parse()

	fmt.Printf("Function: %v\n", os.Getenv("FUNCTION_TARGET"))
	fmt.Printf("Port: %v\n", *port)
	if err := funcframework.Start(*port); err != nil {
		fmt.Println(err)
	}
}
