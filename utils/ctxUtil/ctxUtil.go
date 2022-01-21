package ctxUtil

import (
	"context"
	"google.golang.org/grpc/metadata"
)

func GetHeader(ctx context.Context) metadata.MD {
	// Read metadata from client.
	md, getok := metadata.FromIncomingContext(ctx)
	if !getok {
		return nil
	}
	return md
}
func SetHeader(ctx context.Context, Name, value string) bool {
	// Read metadata from client.
	md, getok := metadata.FromIncomingContext(ctx)
	if !getok {
		return false
	}
	md.Set(Name, value)
	return true
}
