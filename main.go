package main

import (
	"context"
	"fmt"

	"github.com/containers/buildah"
	"github.com/containers/buildah/pkg/parse"
	"github.com/containers/common/pkg/config"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
)

func main() {
	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	buildStoreOptions, err := storage.DefaultStoreOptionsAutoDetectUID()
	if err != nil {
		panic(err)
	}

	conf, err := config.Default()
	if err != nil {
		panic(err)
	}
	capabilitiesForRoot, err := conf.Capabilities("root", nil, nil)
	if err != nil {
		panic(err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		panic(err)
	}
	defer buildStore.Shutdown(false)

	builderOpts := buildah.BuilderOptions{
		FromImage:    "node:12-alpine",
		Capabilities: capabilitiesForRoot,
	}

	builder, err := buildah.NewBuilder(context.TODO(), buildStore, builderOpts)
	if err != nil {
		panic(err)
	}
	defer builder.Delete()

	err = builder.Add("/home/node/", false, buildah.AddAndCopyOptions{}, "script.js")
	if err != nil {
		panic(err)
	}

	isolation, err := parse.IsolationOption("")
	if err != nil {
		panic(err)
	}

	err = builder.Run([]string{"sh", "-c", "date > /home/node/build-date.txt"}, buildah.RunOptions{Isolation: isolation, Terminal: buildah.WithoutTerminal})
	if err != nil {
		panic(err)
	}

	builder.SetCmd([]string{"node", "/home/node/script.js"})

	imageRef, err := is.Transport.ParseStoreReference(buildStore, "docker.io/myusername/my-image")
	if err != nil {
		panic(err)
	}

	imageId, _, _, err := builder.Commit(context.TODO(), imageRef, buildah.CommitOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image built! %s\n", imageId)
}
