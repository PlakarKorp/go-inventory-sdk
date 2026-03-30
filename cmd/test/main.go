package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/PlakarKorp/go-inventory-sdk/inventory"
	"github.com/PlakarKorp/pkg"

	"github.com/PlakarKorp/integration-aws-inventory"
	"github.com/PlakarKorp/integration-ovh-inventory"
)

type Factory func(context.Context, map[string]string) (inventory.Inventory, error)

func show(v any) string {
	buf, _ := json.MarshalIndent(v, "", "   ")
	return string(buf)
}

func main() {
	cfg := make(map[string]string)
	flag.Parse()

	var factory Factory

	backend := flag.Arg(0)
	switch backend {
	case "ovh-inventory":
		factory = ovhinventory.NewInventory
		cfg["endpoint"] = "ovh-eu"
		cfg["application_key"] = os.Getenv("APPLICATION_KEY")
		cfg["application_secret"] = os.Getenv("APPLICATION_SECRET")
		cfg["consumer_key"] = os.Getenv("CONSUMER_KEY")
	case "aws-inventory":
		factory = awsinventory.NewInventory
		cfg["credentials_type"] = "access_key"
		cfg["access_key"] = os.Getenv("AWS_ACCESS_KEY")
		cfg["secret_access_key"] = os.Getenv("AWS_SECRET_KEY")
		cfg["region"] = "eu-west-3"
	case "aws-iam":
		factory = awsinventory.NewInventory
		cfg["credentials_type"] = "iam"
		backend = "aws-inventory"
	default:
		panic(fmt.Errorf("unknown backend %s", backend))
	}

	ctx := context.Background()
	inv, err := factory(ctx, cfg)
	if err != nil {
		panic(err)
	}

	c := make(chan *inventory.InventoryEntry)
	if err := inv.List(ctx, c); err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-c:
			if !ok {
				return
			}
			if r.Error != nil {
				fmt.Println("Error", r.Error)
				continue
			}
			if r.Class != pkg.ResourceClassUndefined {
				fmt.Println(show(r))
			}
		}
	}
}
