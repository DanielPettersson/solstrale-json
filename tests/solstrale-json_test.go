package tests

import (
	"errors"
	"io/ioutil"
	"testing"

	solstralejson "github.com/DanielPettersson/solstrale-json"
	"github.com/DanielPettersson/solstrale/post"
	"github.com/DanielPettersson/solstrale/renderer"
	"github.com/stretchr/testify/assert"
)

func TestToSceneErrors(t *testing.T) {

	testCases := map[string]error{
		"scene-root-missing-world.json":               errors.New("scene is missing world"),
		"scene-root-missing-camera.json":              errors.New("scene is missing camera"),
		"scene-root-missing-background.json":          errors.New("scene is missing background"),
		"scene-root-missing-renderConfig.json":        errors.New("scene is missing renderConfig"),
		"renderConfig-missing-imageWidth.json":        errors.New("renderConfig is missing imageWidth"),
		"renderConfig-missing-imageHeight.json":       errors.New("renderConfig is missing imageHeight"),
		"renderConfig-missing-samplesPerPixel.json":   errors.New("renderConfig is missing samplesPerPixel"),
		"renderConfig-missing-shader.json":            errors.New("renderConfig is missing shader"),
		"renderConfig-missing-postProcessor.json":     errors.New("renderConfig is missing postProcessor"),
		"bvh-missing-list.json":                       errors.New("bvh is missing list"),
		"bvh-empty-list.json":                         errors.New("bvh has empty list"),
		"constantMedium-missing-object.json":          errors.New("constantMedium is missing object"),
		"constantMedium-invalid-object.json":          errors.New("hittable is missing type"),
		"constantMedium-missing-density.json":         errors.New("constantMedium is missing density"),
		"constantMedium-missing-color.json":           errors.New("constantMedium is missing color"),
		"constantMedium-invalid-color.json":           errors.New("texture is missing type"),
		"hittableList-missing-list.json":              errors.New("hittableList is missing list"),
		"hittableList-empty-list.json":                errors.New("hittableList has empty list"),
		"motionBlur-missing-object.json":              errors.New("motionBlur is missing object"),
		"motionBlur-invalid-object.json":              errors.New("hittable is missing type"),
		"motionBlur-missing-blurDirection.json":       errors.New("motionBlur is missing blurDirection"),
		"quad-missing-corner.json":                    errors.New("quad is missing corner"),
		"quad-missing-dirU.json":                      errors.New("quad is missing dirU"),
		"quad-missing-dirV.json":                      errors.New("quad is missing dirV"),
		"quad-missing-mat.json":                       errors.New("quad is missing mat"),
		"quad-invalid-mat.json":                       errors.New("material is missing type"),
		"rotationY-missing-object.json":               errors.New("rotationY is missing object"),
		"rotationY-invalid-object.json":               errors.New("hittable is missing type"),
		"rotationY-missing-angle.json":                errors.New("rotationY is missing angle"),
		"sphere-missing-center.json":                  errors.New("sphere is missing center"),
		"sphere-missing-radius.json":                  errors.New("sphere is missing radius"),
		"sphere-missing-mat.json":                     errors.New("sphere is missing mat"),
		"translation-missing-object.json":             errors.New("translation is missing object"),
		"translation-invalid-object.json":             errors.New("hittable is missing type"),
		"translation-missing-offset.json":             errors.New("translation is missing offset"),
		"hittable-missing-type.json":                  errors.New("hittable is missing type"),
		"hittable-invalid-type.json":                  errors.New("unexpected hittable type: monkey"),
		"texture-missing-type.json":                   errors.New("texture is missing type"),
		"texture-invalid-type.json":                   errors.New("unexpected texture type: monkey"),
		"solidColor-missing-color.json":               errors.New("solidColor is missing color"),
		"checker-missing-scale.json":                  errors.New("checker is missing scale"),
		"checker-missing-even.json":                   errors.New("checker is missing even"),
		"checker-invalid-even.json":                   errors.New("texture is missing type"),
		"checker-missing-odd.json":                    errors.New("checker is missing odd"),
		"checker-invalid-odd.json":                    errors.New("texture is missing type"),
		"image-missing-path.json":                     errors.New("image is missing path"),
		"image-exist-path.json":                       errors.New("open abcd: no such file or directory"),
		"image-format-path.json":                      errors.New("image: unknown format"),
		"image-missing-mirror.json":                   errors.New("image is missing mirror"),
		"noise-missing-color.json":                    errors.New("noise is missing color"),
		"noise-missing-scale.json":                    errors.New("noise is missing scale"),
		"material-missing-type.json":                  errors.New("material is missing type"),
		"material-invalid-type.json":                  errors.New("unexpected material type: monkey"),
		"lambertian-missing-texture.json":             errors.New("lambertian is missing texture"),
		"metal-missing-texture.json":                  errors.New("metal is missing texture"),
		"metal-invalid-texture.json":                  errors.New("texture is missing type"),
		"metal-missing-fuzz.json":                     errors.New("metal is missing fuzz"),
		"dielectric-missing-texture.json":             errors.New("dielectric is missing texture"),
		"dielectric-invalid-texture.json":             errors.New("texture is missing type"),
		"dielectric-missing-indexOfRefraction.json":   errors.New("dielectric is missing indexOfRefraction"),
		"diffuseLight-missing-texture.json":           errors.New("diffuseLight is missing texture"),
		"diffuseLight-invalid-texture.json":           errors.New("texture is missing type"),
		"camera-missing-imageWidth.json":              errors.New("camera is missing imageWidth"),
		"camera-missing-imageHeight.json":             errors.New("camera is missing imageHeight"),
		"camera-missing-verticalFovDegrees.json":      errors.New("camera is missing verticalFovDegrees"),
		"camera-missing-apertureSize.json":            errors.New("camera is missing apertureSize"),
		"camera-missing-focusDistance.json":           errors.New("camera is missing focusDistance"),
		"camera-missing-lookFrom.json":                errors.New("camera is missing lookFrom"),
		"camera-missing-lookAt.json":                  errors.New("camera is missing lookAt"),
		"camera-missing-vup.json":                     errors.New("camera is missing vup"),
		"vec-missing-x.json":                          errors.New("vec is missing x"),
		"vec-missing-y.json":                          errors.New("vec is missing y"),
		"vec-missing-z.json":                          errors.New("vec is missing z"),
		"pathTracing-missing-samplesPerPixel.json":    errors.New("pathTracing is missing samplesPerPixel"),
		"shader-missing-type.json":                    errors.New("shader is missing type"),
		"shader-invalid-type.json":                    errors.New("unexpected shader type: monkey"),
		"postProcessor-missing-type.json":             errors.New("postProcessor is missing type"),
		"postProcessor-invalid-type.json":             errors.New("unexpected postProcessor type: monkey"),
		"oidn-missing-oidnDenoiseExecutablePath.json": errors.New("oidn is missing oidnDenoiseExecutablePath"),
		"float-invalid-type.json":                     errors.New("vec expected number type for x"),
		"string-invalid-type.json":                    errors.New("hittable expected string type for type"),
		"object-invalid-type.json":                    errors.New("sphere expected object type for center"),
	}

	for fileName, expectedErr := range testCases {
		scene, err := fileToScene(t, fileName)
		assert.Nil(t, scene)
		assert.Equal(t, expectedErr.Error(), err.Error())
	}
}

func TestToSceneRenderConfig(t *testing.T) {
	scene, err := fileToScene(t, "renderConfig.json")
	assert.Nil(t, err)
	actual := scene.RenderConfig
	assert.Equal(t, renderer.RenderConfig{
		ImageWidth:      1,
		ImageHeight:     2,
		SamplesPerPixel: 3,
		Shader:          renderer.AlbedoShader{},
	}, actual)
}

func TestToSceneRenderConfig2(t *testing.T) {
	scene, err := fileToScene(t, "renderConfig2.json")
	assert.Nil(t, err)
	actual := scene.RenderConfig
	assert.Equal(t, renderer.RenderConfig{
		ImageWidth:      1,
		ImageHeight:     2,
		SamplesPerPixel: 3,
		Shader:          renderer.SimpleShader{},
		PostProcessor: post.OidnPostProcessor{
			OidnDenoiseExecutablePath: "/usr/local/oidn",
		},
	}, actual)
}

func TestToSceneBvh(t *testing.T) {
	_, err := fileToScene(t, "bvh.json")
	assert.Nil(t, err)
}

func TestToSceneHittableList(t *testing.T) {
	_, err := fileToScene(t, "hittableList.json")
	assert.Nil(t, err)
}

func TestToSceneHittables(t *testing.T) {
	_, err := fileToScene(t, "hittables.json")
	assert.Nil(t, err)
}

func TestToSceneTextures(t *testing.T) {
	_, err := fileToScene(t, "textures.json")
	assert.Nil(t, err)
}

func TestToSceneMaterials(t *testing.T) {
	_, err := fileToScene(t, "materials.json")
	assert.Nil(t, err)
}

func fileToScene(t *testing.T, fileName string) (*renderer.Scene, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return solstralejson.ToScene(b)
}
