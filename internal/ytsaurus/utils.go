package ytsaurus

import (
	"context"
	"fmt"

	"go.ytsaurus.tech/yt/go/ypath"
	"go.ytsaurus.tech/yt/go/yt"
)

func GetObjectByID(ctx context.Context, client yt.Client, id string, result any) error {
	p := ypath.Path(fmt.Sprintf("#%s/@", id))
	return client.GetNode(ctx, p, result, nil)
}

func RemoveIfExists(ctx context.Context, client yt.Client, p ypath.Path) error {
	ok, err := client.NodeExists(ctx, p, nil)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return client.RemoveNode(ctx, p, nil)
}
