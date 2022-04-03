package tests

import (
	"errors"
	"io/ioutil"
	"testing"

	solstralejson "github.com/DanielPettersson/solstrale-json"
	"github.com/stretchr/testify/assert"
)

func TestToSceneErrors(t *testing.T) {

	testCases := map[string]error{
		"scene-root-missing-world.json":             errors.New("scene is missing world"),
		"scene-root-missing-camera.json":            errors.New("scene is missing camera"),
		"scene-root-missing-background.json":        errors.New("scene is missing background"),
		"scene-root-missing-renderConfig.json":      errors.New("scene is missing renderConfig"),
		"renderConfig-missing-imageWidth.json":      errors.New("renderConfig is missing imageWidth"),
		"renderConfig-missing-imageHeight.json":     errors.New("renderConfig is missing imageHeight"),
		"renderConfig-missing-samplesPerPixel.json": errors.New("renderConfig is missing samplesPerPixel"),
		"renderConfig-missing-shader.json":          errors.New("renderConfig is missing shader"),
		"renderConfig-missing-postProcessor.json":   errors.New("renderConfig is missing postProcessor"),
		"bvh-missing-list.json":                     errors.New("bvh is missing list"),
		"bvh-empty-list.json":                       errors.New("bvh has empty list"),
		"constantMedium-missing-boundary.json":      errors.New("constantMedium is missing boundary"),
		"constantMedium-missing-density.json":       errors.New("constantMedium is missing density"),
		"constantMedium-missing-color.json":         errors.New("constantMedium is missing color"),
		"hittableList-missing-list.json":            errors.New("hittableList is missing list"),
		"hittableList-empty-list.json":              errors.New("hittableList has empty list"),
		"motionBlur-missing-object.json":            errors.New("motionBlur is missing object"),
		"motionBlur-missing-blurDirection.json":     errors.New("motionBlur is missing blurDirection"),
		"quad-missing-corner.json":                  errors.New("quad is missing corner"),
		"quad-missing-dirU.json":                    errors.New("quad is missing dirU"),
		"quad-missing-dirV.json":                    errors.New("quad is missing dirV"),
		"quad-missing-mat.json":                     errors.New("quad is missing mat"),
		"rotationY-missing-object.json":             errors.New("rotationY is missing object"),
		"rotationY-missing-angle.json":              errors.New("rotationY is missing angle"),
		"sphere-missing-center.json":                errors.New("sphere is missing center"),
		"sphere-missing-radius.json":                errors.New("sphere is missing radius"),
		"sphere-missing-mat.json":                   errors.New("sphere is missing mat"),
		"translation-missing-object.json":           errors.New("translation is missing object"),
		"translation-missing-offset.json":           errors.New("translation is missing offset"),
		"hittable-missing-type.json":                errors.New("hittable is missing type"),
		"hittable-invalid-type.json":                errors.New("unexpected hittable type: monkey"),
		"texture-missing-type.json":                 errors.New("texture is missing type"),
		"texture-invalid-type.json":                 errors.New("unexpected texture type: monkey"),
		"solidColor-missing-color.json":             errors.New("solidColor is missing color"),
		"checker-missing-scale.json":                errors.New("checker is missing scale"),
		"checker-missing-even.json":                 errors.New("checker is missing even"),
		"checker-missing-odd.json":                  errors.New("checker is missing odd"),
		"image-missing-path.json":                   errors.New("image is missing path"),
		"image-exist-path.json":                     errors.New("open abcd: no such file or directory"),
		"image-format-path.json":                    errors.New("image: unknown format"),
		"image-missing-mirror.json":                 errors.New("image is missing mirror"),
		"noise-missing-color.json":                  errors.New("noise is missing color"),
		"noise-missing-scale.json":                  errors.New("noise is missing scale"),
		"material-missing-type.json":                errors.New("material is missing type"),
		"material-invalid-type.json":                errors.New("unexpected material type: monkey"),
	}

	for fileName, expectedErr := range testCases {

		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		scene, err := solstralejson.ToScene(b)
		assert.Nil(t, scene)
		assert.Equal(t, expectedErr.Error(), err.Error())
	}
}
