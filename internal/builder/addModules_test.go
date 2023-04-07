package builder_test

import (
	"os"
	"runtime/pprof"
	"testing"
	"bytes"

	"github.com/buildpacks/imgutil/fakes"
	"github.com/buildpacks/lifecycle/api"

	"github.com/buildpacks/pack/pkg/dist"
	"github.com/buildpacks/pack/internal/builder"
	"github.com/buildpacks/pack/pkg/logging"
	ifakes "github.com/buildpacks/pack/internal/fakes"
	"github.com/buildpacks/pack/pkg/buildpack"
)

func BenchmarkAndProfileAddModules(b *testing.B) {
	f, err := os.Create("addModulesCPU.prof")
	if err != nil {
		panic("aw geez")
	}
	defer f.Close()

	if err = pprof.StartCPUProfile(f); err != nil {
		panic("couldn't start profile")
	}
	defer pprof.StopCPUProfile()

	outBuf := bytes.Buffer{}
	logger := logging.NewLogWithWriters(&outBuf, &outBuf)
	baseImage := fakes.NewImage("base/image", "", nil)
	baseImage.SetEnv("CNB_USER_ID", "1234")
	baseImage.SetEnv("CNB_GROUP_ID", "4321")
	baseImage.SetLabel("io.buildpacks.stack.id", "some.stack.id")
	baseImage.SetLabel("io.buildpacks.stack.mixins", `["mixinX", "mixinY", "build:mixinA"]`)
	subject, err := builder.New(baseImage, "some/builder")
	tmpDir, err := os.MkdirTemp("", "create-builder-scratch")

	bp1v1, err := ifakes.NewFakeBuildpack(dist.BuildpackDescriptor{
		WithAPI: api.MustParse("0.2"),
		WithInfo: dist.ModuleInfo{
			ID:      "some.buildpack.id",
			Version: "some.buildpack.version",
		},
		WithStacks: []dist.Stack{{
			ID:     "*",
			Mixins: []string{"mixinA", "build:mixinB", "run:mixinD"},
		}},
	}, 0644)
	subject.AddBuildpack(bp1v1)

	bpLayers := dist.ModuleLayers{}
	if _, err := dist.GetLabel(subject.Image(), dist.BuildpackLayersLabel, &bpLayers); err != nil {
		panic("we failed the unit test we were benchmarking")
	}

	for n := 0; n < b.N; n++ {
		if builder.AddModules(subject, buildpack.KindBuildpack, logger, tmpDir, subject.Image(), subject.AdditionalBuildpacks(), bpLayers) != nil {
			panic("we failed the unit test we were benchmarking")
		}
	}
}


