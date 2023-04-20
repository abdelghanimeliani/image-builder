package main

import (
	"context"
	"errors"
	"os"

	"github.com/containers/buildah"
	"github.com/containers/common/pkg/config"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/sirupsen/logrus"
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

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		panic(err)
	}
	println("this is the buildstore object", buildStore)
	defer buildStore.Shutdown(false)

	conf, err := config.Default()
	if err != nil {
		panic(err)
	}
	capabilitiesForRoot, err := conf.Capabilities("root", nil, nil)
	if err != nil {
		panic(err)
	}
	// Create storage reference
	imageRef, err := is.Transport.ParseStoreReference(buildStore, "localhost/restore-restore")
	if err != nil {

		panic(errors.New("failed to parse image name"))
	}

	// Build an image scratch
	builderOptions := buildah.BuilderOptions{
		FromImage:    "scratch",
		Capabilities: capabilitiesForRoot,
	}
	importBuilder, err := buildah.NewBuilder(context.TODO(), buildStore, builderOptions)
	if err != nil {
		panic(err)
	}
	// Clean up buildah working container
	defer func() {
		if err := importBuilder.Delete(); err != nil {
			logrus.Errorf("Image builder delete failed: %v", err)
		}
	}()

	// Export checkpoint into temporary tar file
	tmpDir, err := os.MkdirTemp("", "checkpoint_image_")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy checkpoint from temporary tar file in the image
	addAndCopyOptions := buildah.AddAndCopyOptions{}
	if err := importBuilder.Add("", true, addAndCopyOptions, "checkpoint.tar"); err != nil {
		panic(err)
	}

	importBuilder.SetAnnotation("io.kubernetes.cri-o.annotations.checkpoint.name", "counter")
	commitOptions := buildah.CommitOptions{
		Squash:        true,
		SystemContext: &types.SystemContext{},
	}

	// Create checkpoint image
	id, _, _, err := importBuilder.Commit(context.TODO(), imageRef, commitOptions)
	if err != nil {
		panic(err)
	}
	logrus.Debugf("Created checkpoint image: %s", id)
}
