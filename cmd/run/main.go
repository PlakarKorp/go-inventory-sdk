package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PlakarKorp/go-inventory-sdk/inventory"
	sdk "github.com/PlakarKorp/go-inventory-sdk/sdk/client"
	"github.com/PlakarKorp/pkg"
)

func main() {
	var opt_conffile string
	flag.StringVar(&opt_conffile, "config", "", "config file")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Missing executable")
		os.Exit(1)
	}

	config := make(map[string]string)

	if opt_conffile != "" {
		file, err := os.Open(opt_conffile)
		if err != nil {
			fmt.Println("os.Open:", err)
			os.Exit(1)
		}

		var conf map[string]string
		err = json.NewDecoder(file).Decode(&conf)
		if err != nil {
			fmt.Println("os.Open:", err)
			os.Exit(1)
		}

		for key, value := range conf {
			switch {
			case strings.HasPrefix(value, "env:"):
				value = os.Getenv(value[4:])
			case strings.HasPrefix(value, "cmd:"):
				out, err := exec.Command("/bin/sh", "-c", value[4:]).CombinedOutput()
				if err != nil {
					fmt.Println("exec.Command:", err)
					os.Exit(1)
				}
				value = strings.TrimRight(string(out), "\r\n")
			case strings.HasPrefix(value, "val:"):
				value = value[4:]
			default:
			}
			config[key] = value
		}
	}

	ctx := context.Background()
	bin := flag.Arg(0)
	args := flag.Args()[1:]
	client, conn, err := sdk.ConnectPlugin(ctx, bin, args)
	if err != nil {
		fmt.Println("ConnectPlugin:", err)
		os.Exit(1)
	}

	inv, err := sdk.NewGrpcInventory(ctx, client, config)
	if err != nil {
		fmt.Println("NewGrpcInventory:", err)
		os.Exit(1)
	}

	run(ctx, inv)

	err = conn.Close()
	if err != nil {
		fmt.Println("conn.Close:", err)
		os.Exit(1)
	}
}

func show(v any) string {
	buf, _ := json.MarshalIndent(v, "", "   ")
	return string(buf)
}

func run(ctx context.Context, inv inventory.Inventory) {

	c := make(chan *inventory.InventoryEntry)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for r := range c {
			if r.Class != pkg.ResourceClassUndefined {
				fmt.Println(show(r))
			}
		}
	}()

	if err := inv.List(ctx, c); err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	<-done
}
