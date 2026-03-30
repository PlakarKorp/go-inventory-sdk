package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/PlakarKorp/go-inventory-sdk/inventory"
	"google.golang.org/grpc"
)

type InventoryFn func(context.Context, map[string]string) (inventory.Inventory, error)

type grpcInventory struct {
	UnimplementedInventoryServer

	inventory   inventory.Inventory
	constructor InventoryFn
}

func (g *grpcInventory) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	inventory, err := g.constructor(ctx, req.Config)
	if err != nil {
		return nil, err
	}

	g.inventory = inventory
	return &InitResponse{}, nil
}

func (g *grpcInventory) List(req *ListRequest, stream grpc.ServerStreamingServer[ListResponse]) error {
	return errors.ErrUnsupported
}

func RunInventory(constructor InventoryFn) error {
	conn, listener, err := InitConn()
	if err != nil {
		return fmt.Errorf("failed to initialize connection: %w", err)
	}
	defer conn.Close()

	return RunImporterOn(constructor, listener)
}

func RunImporterOn(constructor InventoryFn, listener net.Listener) error {
	server := grpc.NewServer()
	RegisterInventoryServer(server, &grpcInventory{
		constructor: constructor,
	})
	if err := server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func Entrypoint(args []string, constructor InventoryFn) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", args[0])
		os.Exit(1)
	}

	if err := RunInventory(constructor); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(os.Stderr, "Inventory plugin failed unexpectedly: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
