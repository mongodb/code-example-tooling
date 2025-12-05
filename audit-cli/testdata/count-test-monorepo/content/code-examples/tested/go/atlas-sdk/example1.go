// Atlas Go SDK example 1
package main

import (
	"context"
	"go.mongodb.org/atlas-sdk/v20231115002/admin"
)

func main() {
	sdk, _ := admin.NewClient()
	_ = sdk
}

