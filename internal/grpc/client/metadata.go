package client

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func withAuthorization(ctx context.Context, authorization string) context.Context {
	if authorization == "" {
		return ctx
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	md.Set("authorization", authorization)

	return metadata.NewOutgoingContext(ctx, md)
}
