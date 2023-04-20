package main

import (
	"context"
	"fmt"

	"github.com/containers/buildah"
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
		Capabilities: capabilitiesForRoot,
	}

	builder, err := buildah.NewBuilder(context.TODO(), buildStore, builderOpts)
	if err != nil {
		panic(err)
	}
	defer builder.Delete()

	err = builder.Add("/", false, buildah.AddAndCopyOptions{}, "/var/lib/kubelet/checkpoints/checkpoint-counters_default-counter-2023-04-17T11:55:18Z.tar")
	if err != nil {
		panic(err)
	}

	// isolation, err := parse.IsolationOption("")
	// if err != nil {
	// 	panic(err)
	// }

	builder.SetAnnotation("io.kubernetes.cri-o.annotations.checkpoint.name", "counter")

	fmt.Println("this is container annotations:", builder.Annotations())

	// err = builder.Run([]string{"sh", "-c", "date > /home/node/build-date.txt"}, buildah.RunOptions{Isolation: isolation, Terminal: buildah.WithoutTerminal})
	// if err != nil {
	// 	panic(err)
	// }

	// builder.SetCmd([]string{"node", "/home/node/script.js"})

	imageRef, err := is.Transport.ParseStoreReference(buildStore, "localhost/restore-restore")
	if err != nil {
		panic(err)
	}

	imageId, _, _, err := builder.Commit(context.TODO(), imageRef, buildah.CommitOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image built! %s\n", imageId)
}
