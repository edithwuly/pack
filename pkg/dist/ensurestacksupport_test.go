package dist_test

import (
	"os"
	"runtime/pprof"
	"testing"

	"github.com/buildpacks/pack/pkg/dist"
)

func BenchmarkAndProfileEnsureStackSupport(b *testing.B) {
	f, err := os.Create("ensureStackCPU.prof")
	if err != nil {
		panic("aw geez")
	}
	defer f.Close()

	if err = pprof.StartCPUProfile(f); err != nil {
		panic("couldn't start profile")
	}
	defer pprof.StopCPUProfile()

	bp := dist.BuildpackDescriptor{
		WithInfo: dist.ModuleInfo{
			ID:      "some.buildpack.id",
			Version: "some.buildpack.version",
		},
		WithStacks: []dist.Stack{{
			ID:     "some.stack.id",
			Mixins: []string{"mixinA", "build:mixinB", "run:mixinD"},
		}},
	}

	providedMixins := []string{"mixinA", "build:mixinB", "mixinC"}

	for n := 0; n < b.N; n++ {
		if bp.EnsureStackSupport("some.stack.id", providedMixins, false) != nil {
			panic("we failed the unit test we were benchmarking")
		}
	}

	bp = dist.BuildpackDescriptor{
		WithInfo: dist.ModuleInfo{
			ID:      "some.buildpack.id",
			Version: "some.buildpack.version",
		},
		WithStacks: []dist.Stack{{
			ID:     "*",
			Mixins: []string{"mixinA", "build:mixinB", "run:mixinD"},
		}},
	}

	providedMixins = []string{"mixinA", "build:mixinB", "mixinC"}
	for n := 0; n < b.N; n++ {
		if bp.EnsureStackSupport("some.stack.id", providedMixins, false) != nil {
			panic("we failed the unit test")
		}
	}
	/*
				it("returns an error with any missing (and non-ignored) mixins", func() {
					bp := dist.BuildpackDescriptor{
						WithInfo: dist.ModuleInfo{
							ID:      "some.buildpack.id",
							Version: "some.buildpack.version",
						},
						WithStacks: []dist.Stack{{
							ID:     "some.stack.id",
							Mixins: []string{"mixinX", "mixinY", "run:mixinZ"},
						}},
					}

					providedMixins := []string{"mixinA", "mixinB"}
					err := bp.EnsureStackSupport("some.stack.id", providedMixins, false)

					h.AssertError(t, err, "buildpack 'some.buildpack.id@some.buildpack.version' requires missing mixin(s): mixinX, mixinY")
				})
			})

			when("validating against run image mixins", func() {
				it("requires run-only mixins", func() {
					bp := dist.BuildpackDescriptor{
						WithInfo: dist.ModuleInfo{
							ID:      "some.buildpack.id",
							Version: "some.buildpack.version",
						},
						WithStacks: []dist.Stack{{
							ID:     "some.stack.id",
							Mixins: []string{"mixinA", "build:mixinB", "run:mixinD"},
						}},
					}

					providedMixins := []string{"mixinA", "build:mixinB", "mixinC", "run:mixinD"}

					h.AssertNil(t, bp.EnsureStackSupport("some.stack.id", providedMixins, true))
				})

				it("returns an error with any missing mixins", func() {
					bp := dist.BuildpackDescriptor{
						WithInfo: dist.ModuleInfo{
							ID:      "some.buildpack.id",
							Version: "some.buildpack.version",
						},
						WithStacks: []dist.Stack{{
							ID:     "some.stack.id",
							Mixins: []string{"mixinX", "mixinY", "run:mixinZ"},
						}},
					}

					providedMixins := []string{"mixinA", "mixinB"}

					err := bp.EnsureStackSupport("some.stack.id", providedMixins, true)

					h.AssertError(t, err, "buildpack 'some.buildpack.id@some.buildpack.version' requires missing mixin(s): mixinX, mixinY, run:mixinZ")
				})
			})

			it("returns an error when buildpack does not support stack", func() {
				bp := dist.BuildpackDescriptor{
					WithInfo: dist.ModuleInfo{
						ID:      "some.buildpack.id",
						Version: "some.buildpack.version",
					},
					WithStacks: []dist.Stack{{
						ID:     "some.stack.id",
						Mixins: []string{"mixinX", "mixinY"},
					}},
				}

				err := bp.EnsureStackSupport("some.nonexistent.stack.id", []string{"mixinA"}, true)

				h.AssertError(t, err, "buildpack 'some.buildpack.id@some.buildpack.version' does not support stack 'some.nonexistent.stack.id")
			})

			it("skips validating order buildpack", func() {
				bp := dist.BuildpackDescriptor{
					WithInfo: dist.ModuleInfo{
						ID:      "some.buildpack.id",
						Version: "some.buildpack.version",
					},
					WithStacks: []dist.Stack{},
				}

				h.AssertNil(t, bp.EnsureStackSupport("some.stack.id", []string{"mixinA"}, true))
			})
		})
	*/
}
