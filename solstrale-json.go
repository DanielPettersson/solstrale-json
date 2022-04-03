package solstralejson

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"reflect"

	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/post"
	"github.com/DanielPettersson/solstrale/renderer"
)

func ToScene(jsonBytes []byte) (*renderer.Scene, error) {
	var data map[string]interface{}
	json.Unmarshal(jsonBytes, &data)

	return toScene(data)
}

func toScene(data map[string]interface{}) (*renderer.Scene, error) {

	worldData, err := getObject("scene", data, "world")
	if err != nil {
		return nil, err
	}
	world, err := toHittable(worldData)
	if err != nil {
		return nil, err
	}

	cameraData, err := getObject("scene", data, "camera")
	if err != nil {
		return nil, err
	}
	camera, err := toCamera(cameraData)
	if err != nil {
		return nil, err
	}

	background, err := getVec("scene", data, "background")
	if err != nil {
		return nil, err
	}

	renderConfigData, err := getObject("scene", data, "renderConfig")
	if err != nil {
		return nil, err
	}
	renderConfig, err := toRenderConfig(renderConfigData)
	if err != nil {
		return nil, err
	}

	return &renderer.Scene{
		World:           world,
		Cam:             *camera,
		BackgroundColor: *background,
		RenderConfig:    *renderConfig,
	}, nil
}

func toRenderConfig(data map[string]interface{}) (*renderer.RenderConfig, error) {
	imageWidth, err := getFloat("renderConfig", data, "imageWidth")
	if err != nil {
		return nil, err
	}

	imageHeight, err := getFloat("renderConfig", data, "imageHeight")
	if err != nil {
		return nil, err
	}

	samplesPerPixel, err := getFloat("renderConfig", data, "samplesPerPixel")
	if err != nil {
		return nil, err
	}

	shaderData, err := getObject("renderConfig", data, "shader")
	if err != nil {
		return nil, err
	}
	shader, err := toShader(shaderData)
	if err != nil {
		return nil, err
	}

	postProcessorData, err := getObject("renderConfig", data, "postProcessor")
	if err != nil {
		return nil, err
	}
	postProcessor, err := toPostProcessor(postProcessorData)
	if err != nil {
		return nil, err
	}

	return &renderer.RenderConfig{
		ImageWidth:      int(imageWidth),
		ImageHeight:     int(imageHeight),
		SamplesPerPixel: int(samplesPerPixel),
		Shader:          shader,
		PostProcessor:   postProcessor,
	}, nil
}

func toShader(data map[string]interface{}) (renderer.Shader, error) {
	t, err := getString("shader", data, "type")
	if err != nil {
		return nil, err
	}

	switch t {
	case "pathTracing":
		return toPathTracing(data)
	case "albedo":
		return renderer.AlbedoShader{}, nil
	case "normal":
		return renderer.NormalShader{}, nil
	case "simple":
		return renderer.SimpleShader{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected hittable type: %v", t))
	}
}

func toPathTracing(data map[string]interface{}) (renderer.Shader, error) {
	samplesPerPixel, err := getFloat("pathTracing", data, "samplesPerPixel")
	if err != nil {
		return nil, err
	}

	return renderer.PathTracingShader{
		MaxDepth: int(samplesPerPixel),
	}, nil
}

func toPostProcessor(data map[string]interface{}) (post.PostProcessor, error) {
	if data == nil {
		return nil, nil
	} else {
		t, err := getString("postProcessor", data, "type")
		if err != nil {
			return nil, err
		}

		switch t {
		case "oidn":
			return toOidn(data)
		default:
			return nil, errors.New(fmt.Sprintf("Unexpected hittable type: %v", t))
		}
	}
}

func toOidn(data map[string]interface{}) (post.PostProcessor, error) {
	oidnDenoiseExecutablePath, err := getString("hittable", data, "oidnDenoiseExecutablePath")
	if err != nil {
		return nil, err
	}
	return post.OidnPostProcessor{
		OidnDenoiseExecutablePath: oidnDenoiseExecutablePath,
	}, nil
}

func toHittable(data map[string]interface{}) (hittable.Hittable, error) {
	t, err := getString("hittable", data, "type")
	if err != nil {
		return nil, err
	}

	switch t {
	case "bvh":
		return toBvh(data)
	case "constantMedium":
		return toConstantMedium(data)
	case "hittableList":
		return toHittableList(data)
	case "motionBlur":
		return toMotionBlur(data)
	case "quad":
		return toQuad(data)
	case "rotationY":
		return toRotationY(data)
	case "sphere":
		return toSphere(data)
	case "translation":
		return toTranslation(data)
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected hittable type: %v", t))
	}
}

func toBvh(data map[string]interface{}) (hittable.Hittable, error) {
	list, err := getAttr("bvh", data, "list")
	if err != nil {
		return nil, err
	}

	items := hittable.NewHittableList()
	for _, itemData := range list.([]interface{}) {
		item, err := toObject("bvh", itemData, "item")
		if err != nil {
			return nil, err
		}
		hittable, err := toHittable(item)
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
	}
	if len(items.List()) == 0 {
		return nil, errors.New("bvh has empty list")
	}

	bvh := hittable.NewBoundingVolumeHierarchy(items)
	return bvh, nil
}

func toConstantMedium(data map[string]interface{}) (hittable.Hittable, error) {
	boundaryData, err := getObject("constantMedium", data, "boundary")
	if err != nil {
		return nil, err
	}
	boundary, err := toHittable(boundaryData)
	if err != nil {
		return nil, err
	}

	density, err := getFloat("constantMedium", data, "density")
	if err != nil {
		return nil, err
	}

	colorData, err := getObject("constantMedium", data, "color")
	if err != nil {
		return nil, err
	}
	color, err := toTexture(colorData)
	if err != nil {
		return nil, err
	}

	constantMedium := hittable.NewConstantMedium(
		boundary,
		density,
		color,
	)

	return constantMedium, nil
}

func toHittableList(data map[string]interface{}) (hittable.Hittable, error) {
	list, err := getAttr("hittableList", data, "list")
	if err != nil {
		return nil, err
	}

	items := hittable.NewHittableList()
	for _, itemData := range list.([]interface{}) {
		item, err := toObject("hittableList", itemData, "item")
		if err != nil {
			return nil, err
		}
		hittable, err := toHittable(item)
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
	}
	if len(items.List()) == 0 {
		return nil, errors.New("hittableList has empty list")
	}

	return &items, nil
}

func toMotionBlur(data map[string]interface{}) (hittable.Hittable, error) {
	objectData, err := getObject("motionBlur", data, "object")
	if err != nil {
		return nil, err
	}
	object, err := toHittable(objectData)
	if err != nil {
		return nil, err
	}

	blurDirection, err := getVec("motionBlur", data, "blurDirection")
	if err != nil {
		return nil, err
	}

	motionBlur := hittable.NewMotionBlur(
		object,
		*blurDirection,
	)

	return motionBlur, nil
}

func toQuad(data map[string]interface{}) (hittable.Hittable, error) {
	corner, err := getVec("quad", data, "corner")
	if err != nil {
		return nil, err
	}

	dirU, err := getVec("quad", data, "dirU")
	if err != nil {
		return nil, err
	}

	dirV, err := getVec("quad", data, "dirV")
	if err != nil {
		return nil, err
	}

	matData, err := getObject("quad", data, "mat")
	if err != nil {
		return nil, err
	}
	mat, err := toMaterial(matData)
	if err != nil {
		return nil, err
	}

	quad := hittable.NewQuad(
		*corner,
		*dirU,
		*dirV,
		mat,
	)

	return quad, nil
}

func toRotationY(data map[string]interface{}) (hittable.Hittable, error) {
	objectData, err := getObject("rotationY", data, "object")
	if err != nil {
		return nil, err
	}
	object, err := toHittable(objectData)
	if err != nil {
		return nil, err
	}

	angle, err := getFloat("rotationY", data, "angle")
	if err != nil {
		return nil, err
	}

	rotationY := hittable.NewRotationY(
		object,
		angle,
	)

	return rotationY, nil
}

func toSphere(data map[string]interface{}) (hittable.Hittable, error) {
	center, err := getVec("sphere", data, "center")
	if err != nil {
		return nil, err
	}

	radius, err := getFloat("sphere", data, "radius")
	if err != nil {
		return nil, err
	}

	matData, err := getObject("sphere", data, "mat")
	if err != nil {
		return nil, err
	}
	mat, err := toMaterial(matData)
	if err != nil {
		return nil, err
	}

	sphere := hittable.NewSphere(
		*center,
		radius,
		mat,
	)

	return sphere, nil
}

func toTranslation(data map[string]interface{}) (hittable.Hittable, error) {
	objectData, err := getObject("translation", data, "object")
	if err != nil {
		return nil, err
	}
	object, err := toHittable(objectData)
	if err != nil {
		return nil, err
	}

	offset, err := getVec("translation", data, "offset")
	if err != nil {
		return nil, err
	}

	translation := hittable.NewTranslation(
		object,
		*offset,
	)

	return translation, nil
}

func toTexture(data map[string]interface{}) (material.Texture, error) {
	t, err := getString("texture", data, "type")
	if err != nil {
		return nil, err
	}

	switch t {
	case "solidColor":
		return toSolidColor(data)
	case "checker":
		return toChecker(data)
	case "image":
		return toImage(data)
	case "noise":
		return toNoise(data)
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected texture type: %v", t))
	}
}

func toSolidColor(data map[string]interface{}) (material.Texture, error) {
	color, err := getVec("solidColor", data, "color")
	if err != nil {
		return nil, err
	}

	return material.SolidColor{
		ColorValue: *color,
	}, nil
}

func toChecker(data map[string]interface{}) (material.Texture, error) {
	scale, err := getFloat("checker", data, "scale")
	if err != nil {
		return nil, err
	}

	evenData, err := getObject("checker", data, "even")
	if err != nil {
		return nil, err
	}
	even, err := toTexture(evenData)
	if err != nil {
		return nil, err
	}

	oddData, err := getObject("checker", data, "odd")
	if err != nil {
		return nil, err
	}
	odd, err := toTexture(oddData)
	if err != nil {
		return nil, err
	}

	return material.CheckerTexture{
		Scale: scale,
		Even:  even,
		Odd:   odd,
	}, nil
}

func toImage(data map[string]interface{}) (material.Texture, error) {
	pathData, err := getString("image", data, "path")
	if err != nil {
		return nil, err
	}

	f, err := os.Open(pathData)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	mirror, err := getAttr("image", data, "mirror")
	if err != nil {
		return nil, err
	}

	imageTexture := material.ImageTexture{
		Image:  image,
		Mirror: mirror.(bool),
	}
	return imageTexture, nil
}

func toNoise(data map[string]interface{}) (material.Texture, error) {
	color, err := getVec("noise", data, "color")
	if err != nil {
		return nil, err
	}

	scale, err := getFloat("noise", data, "scale")
	if err != nil {
		return nil, err
	}

	noiseTexture := material.NoiseTexture{
		ColorValue: *color,
		Scale:      scale,
	}

	return noiseTexture, nil
}

func toMaterial(data map[string]interface{}) (material.Material, error) {
	t, err := getString("material", data, "type")
	if err != nil {
		return nil, err
	}

	switch t {
	case "lambertian":
		return toLambertian(data)
	case "metal":
		return toMetal(data)
	case "dielectric":
		return toDielectric(data)
	case "diffuseLight":
		return toDiffuseLight(data)
	case "isotropic":
		return toIsotropic(data)
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected material type: %v", t))
	}
}

func toLambertian(data map[string]interface{}) (material.Material, error) {
	textureData, err := getObject("lambertian", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData)
	if err != nil {
		return nil, err
	}

	lambertian := material.Lambertian{
		Tex: texture,
	}
	return lambertian, nil
}

func toMetal(data map[string]interface{}) (material.Material, error) {
	textureData, err := getObject("metal", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData)
	if err != nil {
		return nil, err
	}

	fuzz, err := getFloat("metal", data, "fuzz")
	if err != nil {
		return nil, err
	}

	metal := material.Metal{
		Tex:  texture,
		Fuzz: fuzz,
	}
	return metal, nil
}

func toDielectric(data map[string]interface{}) (material.Material, error) {
	textureData, err := getObject("dielectric", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData)
	if err != nil {
		return nil, err
	}

	indexOfRefraction, err := getFloat("dielectric", data, "indexOfRefraction")
	if err != nil {
		return nil, err
	}

	metal := material.Dielectric{
		Tex:               texture,
		IndexOfRefraction: indexOfRefraction,
	}
	return metal, nil
}

func toDiffuseLight(data map[string]interface{}) (material.Material, error) {
	textureData, err := getObject("diffuseLight", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData)
	if err != nil {
		return nil, err
	}

	diffuseLight := material.DiffuseLight{
		Emit: texture,
	}
	return diffuseLight, err
}

func toIsotropic(data map[string]interface{}) (material.Material, error) {
	textureData, err := getObject("isotropic", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData)
	if err != nil {
		return nil, err
	}

	isotropic := material.Isotropic{
		Albedo: texture,
	}
	return isotropic, err
}

func toCamera(data map[string]interface{}) (*camera.Camera, error) {
	imageWidth, err := getFloat("camera", data, "imageWidth")
	if err != nil {
		return nil, err
	}

	imageHeight, err := getFloat("camera", data, "imageHeight")
	if err != nil {
		return nil, err
	}

	verticalFovDegrees, err := getFloat("camera", data, "verticalFovDegrees")
	if err != nil {
		return nil, err
	}

	apertureSize, err := getFloat("camera", data, "apertureSize")
	if err != nil {
		return nil, err
	}

	focusDistance, err := getFloat("camera", data, "focusDistance")
	if err != nil {
		return nil, err
	}

	lookFrom, err := getVec("camera", data, "lookFrom")
	if err != nil {
		return nil, err
	}

	lookAt, err := getVec("camera", data, "lookAt")
	if err != nil {
		return nil, err
	}

	vup, err := getVec("camera", data, "vup")
	if err != nil {
		return nil, err
	}

	cam := camera.New(
		int(imageWidth),
		int(imageHeight),
		verticalFovDegrees,
		apertureSize,
		focusDistance,
		*lookFrom,
		*lookAt,
		*vup,
	)
	return &cam, nil
}

func getVec(t string, data map[string]interface{}, key string) (*geo.Vec3, error) {
	vecData, err := getObject(t, data, key)
	if err != nil {
		return nil, err
	}
	vec, err := toVec(vecData)
	if err != nil {
		return nil, err
	}
	return vec, nil
}

func toVec(data map[string]interface{}) (*geo.Vec3, error) {
	x, err := getFloat("vec", data, "x")
	if err != nil {
		return nil, err
	}

	y, err := getFloat("vec", data, "y")
	if err != nil {
		return nil, err
	}

	z, err := getFloat("vec", data, "z")
	if err != nil {
		return nil, err
	}

	vec := geo.NewVec3(x, y, z)
	return &vec, nil
}

func getFloat(t string, data map[string]interface{}, key string) (float64, error) {
	number, err := getAttr(t, data, key)
	if err != nil {
		return 0, err
	}
	if reflect.ValueOf(number).Kind() != reflect.Float64 {
		return 0, errors.New(fmt.Sprintf("%v expected number type for %v", t, key))
	}
	return number.(float64), nil
}

func getString(t string, data map[string]interface{}, key string) (string, error) {
	number, err := getAttr(t, data, key)
	if err != nil {
		return "", err
	}
	if reflect.ValueOf(number).Kind() != reflect.String {
		return "", errors.New(fmt.Sprintf("%v expected string type for %v", t, key))
	}
	return number.(string), nil
}

func getObject(t string, data map[string]interface{}, key string) (map[string]interface{}, error) {
	object, err := getAttr(t, data, key)
	if err != nil {
		return nil, err
	}
	return toObject(t, object, key)
}

func toObject(t string, data interface{}, key string) (map[string]interface{}, error) {
	if reflect.ValueOf(data).Kind() != reflect.Map {
		return nil, errors.New(fmt.Sprintf("%v expected object type for %v", t, key))
	}
	return data.(map[string]interface{}), nil
}

func getAttr(t string, data map[string]interface{}, key string) (interface{}, error) {
	attrVal, ok := data[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%v is missing %v", t, key))
	}
	return attrVal, nil
}
