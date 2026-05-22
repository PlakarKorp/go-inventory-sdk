package server

import (
	"context"

	"github.com/PlakarKorp/go-inventory-sdk/inventory"
	"github.com/PlakarKorp/go-inventory-sdk/sdk"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type InventoryFn func(context.Context, map[string]string) (inventory.Inventory, error)

type grpcInventoryServer struct {
	sdk.UnimplementedInventoryServer

	maxconcurrency int
	inventory      inventory.Inventory
	constructor    InventoryFn
}

func NewGrpcInventoryServer(constructor InventoryFn) *grpcInventoryServer {
	return &grpcInventoryServer{
		constructor: constructor,
	}
}

func (g *grpcInventoryServer) Init(ctx context.Context, req *sdk.InitRequest) (*sdk.InitResponse, error) {
	inventory, err := g.constructor(ctx, req.Config)
	if err != nil {
		return nil, err
	}

	g.inventory = inventory
	return &sdk.InitResponse{}, nil
}

func (g *grpcInventoryServer) List(req *sdk.ListRequest, stream grpc.ServerStreamingServer[sdk.ListResponse]) error {
	entries := make(chan *inventory.InventoryEntry, g.maxconcurrency)
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var wg errgroup.Group

	wg.Go(func() error {
		for res := range entries {
			entry := &sdk.InventoryEntry{
				Class:    string(res.Class),
				Subclass: string(res.SubClass),
				Tags:     res.Tags,
				Urn:      res.URN,
				Name:     res.Name,
				Region:   res.Region,
				Service:  res.Service,
				Resource: res.Resource,
				Details:  res.Details,
			}

			for _, e := range res.Endpoints {
				entry.Endpoints = append(entry.Endpoints, &sdk.HostEndpoint{
					Type:     string(e.Type),
					Endpoint: e.Endpoint,
				})
			}

			if err := stream.Send(&sdk.ListResponse{Entry: entry}); err != nil {
				return err
			}
		}

		return nil
	})

	if err := g.inventory.List(ctx, entries); err != nil {
		return err
	}

	return wg.Wait()
}
