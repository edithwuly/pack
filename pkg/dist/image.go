package dist

import (
	"encoding/json"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/buildpacks/imgutil"

	"github.com/pkg/errors"

	"github.com/buildpacks/pack/internal/style"
)

type Labeled interface {
	Label(name string) (value string, err error)
}

type Labelable interface {
	SetLabel(name string, value string) error
}

func SetLabel(labelable Labelable, label string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "marshalling data to JSON for label %s", style.Symbol(label))
	}
	if err := labelable.SetLabel(label, string(dataBytes)); err != nil {
		return errors.Wrapf(err, "setting label %s", style.Symbol(label))
	}
	return nil
}

func GetLabel(labeled Labeled, label string, obj interface{}) (ok bool, err error) {
	labelData, err := labeled.Label(label)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		return true, nil
	}
	return false, nil
}

func GetPointerLabel(labeled Labeled, label string, obj interface{}) (ok bool, err error) {
	labelData, err := labeled.Label(label)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), &obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		return true, nil
	}
	return false, nil
}

func GetPointerBuildpacksLabel(logger logging.Logger, labeled *imgutil.Image, label string, obj interface{}) (ok bool, err error) {
	logger.Infof("addr of image: %p", labeled)
	logger.Infof("get buildpacks label IO start")
	labelData, err := (*labeled).Label(label)
	logger.Infof("get buildpacks label IO end")
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		return true, nil
	}
	return false, nil
}

func GetBuildpacksLabel(logger logging.Logger, labeled Labeled, label string, obj interface{}) (ok bool, err error) {
	logger.Infof("get buildpacks label IO start")
	labelData, err := labeled.Label(label)
	logger.Infof("get buildpacks label IO end")
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), &obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		return true, nil
	}
	return false, nil
}
