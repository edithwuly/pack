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
	"github.com/buildpacks/pack/pkg/dist"
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

	appPath := "../../../buildpacks/samples/apps/bash-script/"
	descriptorPath := "../../../buildpacks/samples/apps/bash-script/project.toml"
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

func BenchmarkAndProfileFetchBuildpacks(b *testing.B) {
	f, err := os.Create("fetchBuildpacksCPU.prof")
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

	appPath := "../../../buildpacks/samples/apps/bash-script/"
	descriptorPath := "../../../buildpacks/samples/apps/bash-script/project.toml"
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

	relativeBaseDir := opts.RelativeBaseDir
	declaredBPs := opts.Buildpacks
	//fmt.Printf("number of opts.ProjectDescriptor.Build.Buildpacks: %d\n", len(opts.ProjectDescriptor.Build.Buildpacks))
	if len(declaredBPs) == 0 && len(opts.ProjectDescriptor.Build.Buildpacks) != 0 {
		relativeBaseDir = opts.ProjectDescriptorBaseDir

		for _, bp := range opts.ProjectDescriptor.Build.Buildpacks {
			buildpackLocator, err := getBuildpackLocator(bp, bldr.StackID)
			if err != nil {
				panic("we failed the unit test we were benchmarking")
			}
			declaredBPs = append(declaredBPs, buildpackLocator)
		}
	}

	builderBPs := bldr.Buildpacks()
	builderImage := bldr.Image()
	builderOrder := bldr.Order()

	for n := 0; n < b.N; n++ {
		order := dist.Order{{Group: []dist.ModuleRef{}}}
		fetchedBPs := []buildpack.BuildModule{}
		for _, bp := range declaredBPs {
			locatorType, err := buildpack.GetLocatorType(bp, relativeBaseDir, builderBPs)
			if err != nil {
				panic("we failed the unit test we were benchmarking")
			}

			switch locatorType {
			case buildpack.FromBuilderLocator:
				switch {
				case len(order) == 0 || len(order[0].Group) == 0:
					order = builderOrder
				case len(order) > 1:
					// This should only ever be possible if they are using from=builder twice which we don't allow
					panic("we failed the unit test we were benchmarking")
				default:
					newOrder := dist.Order{}
					groupToAdd := order[0].Group
					for _, bOrderEntry := range builderOrder {
						newEntry := dist.OrderEntry{Group: append(groupToAdd, bOrderEntry.Group...)}
						newOrder = append(newOrder, newEntry)
					}

					order = newOrder
				}
			default:
				newFetchedBPs, moduleInfo, err := subject.fetchBuildpack(context.TODO(), bp, relativeBaseDir, builderImage, builderBPs, opts)
				if err != nil {
					panic("we failed the unit test we were benchmarking")
				}
				fetchedBPs = append(fetchedBPs, newFetchedBPs...)
				order = appendBuildpackToOrder(order, *moduleInfo)
			}
		}
	}
}
