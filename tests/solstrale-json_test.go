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
		"scene-root-missing-all.json":                        errors.New("\"world\" value is required\n\"camera\" value is required\n\"background\" value is required\n\"renderConfig\" value is required"),
		"not-json.json":                                      errors.New("error parsing JSON bytes: invalid character 'x' looking for beginning of value"),
		"bvh-with-non-existing-image-path.json":              errors.New("Failed loading image: open : no such file or directory"),
		"bvh-with-wrong-format-image.json":                   errors.New("image: unknown format"),
		"constantMedium-with-non-existing-image-path.json":   errors.New("Failed loading image: open : no such file or directory"),
		"constantMedium-with-non-existing-image-path2.json":  errors.New("Failed loading image: open : no such file or directory"),
		"hittableList-with-non-existing-image-path.json":     errors.New("Failed loading image: open : no such file or directory"),
		"motionBlur-with-non-existing-image-path.json":       errors.New("Failed loading image: open : no such file or directory"),
		"quad-with-non-existing-image-path.json":             errors.New("Failed loading image: open : no such file or directory"),
		"rotationY-with-non-existing-image-path.json":        errors.New("Failed loading image: open : no such file or directory"),
		"sphere-with-non-existing-image-path.json":           errors.New("Failed loading image: open : no such file or directory"),
		"translation-with-non-existing-image-path.json":      errors.New("Failed loading image: open : no such file or directory"),
		"checker-with-non-existing-image-path.json":          errors.New("Failed loading image: open : no such file or directory"),
		"checker2-with-non-existing-image-path.json":         errors.New("Failed loading image: open : no such file or directory"),
		"metal-with-non-existing-image-path.json":            errors.New("Failed loading image: open : no such file or directory"),
		"dielectric-with-non-existing-image-path.json":       errors.New("Failed loading image: open : no such file or directory"),
		"box-with-non-existing-image-path.json":              errors.New("Failed loading image: open : no such file or directory"),
		"triangle-with-non-existing-image-path.json":         errors.New("Failed loading image: open : no such file or directory"),
		"objModel-with-non-existing-image-path.json":         errors.New("Failed loading image: open : no such file or directory"),
		"objModel-with-non-existing-model-path-and-mat.json": errors.New("Failed to read obj file: open missing.obj: no such file or directory"),
		"objModel-with-non-existing-model-path-no-mat.json":  errors.New("Failed to read obj file: open missing.obj: no such file or directory"),
	}

	for fileName, expectedErr := range testCases {

		t.Run(fileName, func(t *testing.T) {
			scene, err := fileToScene(t, fileName)
			assert.Nil(t, scene)
			assert.Equal(t, expectedErr.Error(), err.Error())
		})
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

func TestToSceneRenderConfig3(t *testing.T) {
	scene, err := fileToScene(t, "renderConfig3.json")
	assert.Nil(t, err)
	actual := scene.RenderConfig
	assert.Equal(t, renderer.RenderConfig{
		ImageWidth:      1,
		ImageHeight:     2,
		SamplesPerPixel: 3,
		Shader: renderer.PathTracingShader{
			MaxDepth: 4,
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
