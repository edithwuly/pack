package builder

import (
	"github.com/buildpacks/pack/pkg/buildpack"
)

var AddModules = (*Builder).addModules

func (b *Builder) AdditionalBuildpacks() []buildpack.BuildModule {
	return b.additionalBuildpacks
}