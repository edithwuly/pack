package client

import (
	"os"
	"bytes"
	"context"
	"runtime/pprof"
	"testing"
	"path/filepath"

	dockerclient "github.com/docker/docker/client"

	"github.com/buildpacks/pack/pkg/blob"
	"github.com/buildpacks/pack/pkg/buildpack"
	ifakes "github.com/buildpacks/pack/internal/fakes"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/buildpacks/pack/pkg/project"
	"github.com/buildpacks/pack/pkg/image"
)

func BenchmarkAndProfileProcessBuildpacks(b *testing.B) {
	f, err := os.Create("processBuildpacksCPU.prof")
	if err != nil {
		panic("aw geez")
	}
	defer f.Close()

	if err = pprof.StartCPUProfile(f); err != nil {
		panic("couldn't start profile")
	}
	defer pprof.StopCPUProfile()

	fakeLifecycle := &ifakes.FakeLifecycle{}

	tmpDir, err := os.MkdirTemp("", "build-test")

	docker, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithVersion("1.38"))

	outBuf := bytes.Buffer{}
	logger := logging.NewLogWithWriters(&outBuf, &outBuf)

	dlCacheDir, err := os.MkdirTemp(tmpDir, "dl-cache")
	imageFetcher := image.NewFetcher(logger, docker)
	blobDownloader := blob.NewDownloader(logger, dlCacheDir)
	buildpackDownloader := buildpack.NewDownloader(logger, imageFetcher, blobDownloader, &registryResolver{logger: logger})
	subject := &Client{
		logger:              logger,
		imageFetcher:        imageFetcher,
		downloader:          blobDownloader,
		lifecycleExecutor:   fakeLifecycle,
		docker:              docker,
		buildpackDownloader: buildpackDownloader,
	}

	appPath := "../samples/apps/bash-script/"
	descriptorPath := "../samples/apps/bash-script/project.toml"
	descriptor, err := project.ReadProjectDescriptor(descriptorPath)
	opts := BuildOptions{
		AppPath:           appPath,
		Builder:           "cnbs/sample-builder:jammy",
		Image:             "cnbs/sample-builder:jammy",
		ProjectDescriptorBaseDir: filepath.Dir(descriptorPath),
		ProjectDescriptor:        descriptor,
	}

	builderRef, err := subject.processBuilderName(opts.Builder)

	rawBuilderImage, err := subject.imageFetcher.Fetch(context.TODO(), builderRef.Name(), image.FetchOptions{Daemon: true, PullPolicy: opts.PullPolicy})

	bldr, err := subject.getBuilder(rawBuilderImage)

	for n := 0; n < b.N; n++ {
		if _, _, err := subject.processBuildpacks(context.TODO(), bldr.Image(), bldr.Buildpacks(), bldr.Order(), bldr.StackID, opts); err != nil {
			panic("we failed the unit test we were benchmarking")
		}
	}
}