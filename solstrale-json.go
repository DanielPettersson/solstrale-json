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
	"github.com/DanielPettersson/solstrale/renderer"
)

func ToScene(jsonBytes []byte) (*renderer.Scene, error) {
	var data map[string]interface{}
	json.Unmarshal(jsonBytes, &data)

	return toScene(data)
}

func toScene(data map[string]interface{}) (*renderer.Scene, error) {

	worldData, err := getAttr("scene", data, "world")
	if err != nil {
		return nil, err
	}
	world, err := toHittable(worldData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	cameraData, err := getAttr("scene", data, "camera")
	if err != nil {
		return nil, err
	}
	camera, err := toCamera(cameraData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	background, err := getVec("scene", data, "background")
	if err != nil {
		return nil, err
	}

	return &renderer.Scene{
		World:           world,
		Cam:             *camera,
		BackgroundColor: *background,
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
		return nil, errors.New(fmt.Sprintf("Unexpected hittable type: %v", data))
	}
}

func toBvh(data map[string]interface{}) (hittable.Hittable, error) {
	list, err := getAttr("bvh", data, "list")
	if err != nil {
		return nil, err
	}

	items := hittable.NewHittableList()
	for _, item := range list.([]interface{}) {
		hittable, err := toHittable(item.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
	}
	if len(items.List()) == 0 {
		return nil, errors.New(fmt.Sprintf("bvh has empty list: %v", data))
	}

	bvh := hittable.NewBoundingVolumeHierarchy(items)
	return bvh, nil
}

func toConstantMedium(data map[string]interface{}) (hittable.Hittable, error) {
	boundaryData, err := getAttr("constantMedium", data, "boundary")
	if err != nil {
		return nil, err
	}
	boundary, err := toHittable(boundaryData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	density, err := getFloat("constantMedium", data, "density")
	if err != nil {
		return nil, err
	}

	colorData, err := getAttr("constantMedium", data, "color")
	if err != nil {
		return nil, err
	}
	color, err := toTexture(colorData.(map[string]interface{}))
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
	for _, item := range list.([]interface{}) {
		hittable, err := toHittable(item.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
	}
	if len(items.List()) == 0 {
		return nil, errors.New(fmt.Sprintf("hittableList has empty list: %v", data))
	}

	return &items, nil
}

func toMotionBlur(data map[string]interface{}) (hittable.Hittable, error) {
	blurredHittableData, err := getAttr("motionBlur", data, "blurredHittable")
	if err != nil {
		return nil, err
	}
	blurredHittable, err := toHittable(blurredHittableData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	blurDirection, err := getVec("motionBlur", data, "blurDirection")
	if err != nil {
		return nil, err
	}

	motionBlur := hittable.NewMotionBlur(
		blurredHittable,
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

	matData, err := getAttr("quad", data, "mat")
	if err != nil {
		return nil, err
	}
	mat, err := toMaterial(matData.(map[string]interface{}))
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
	objectData, err := getAttr("rotationY", data, "object")
	if err != nil {
		return nil, err
	}
	object, err := toHittable(objectData.(map[string]interface{}))
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

	matData, err := getAttr("sphere", data, "mat")
	if err != nil {
		return nil, err
	}
	mat, err := toMaterial(matData.(map[string]interface{}))
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
	objectData, err := getAttr("translation", data, "object")
	if err != nil {
		return nil, err
	}
	object, err := toHittable(objectData.(map[string]interface{}))
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
		return nil, errors.New(fmt.Sprintf("Unexpected texture type: %v", data))
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

	evenData, err := getAttr("checker", data, "even")
	if err != nil {
		return nil, err
	}
	even, err := toTexture(evenData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	oddData, err := getAttr("checker", data, "odd")
	if err != nil {
		return nil, err
	}
	odd, err := toTexture(oddData.(map[string]interface{}))
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
		return nil, errors.New(fmt.Sprintf("Unexpected material type: %v", data))
	}
}

func toLambertian(data map[string]interface{}) (material.Material, error) {
	textureData, err := getAttr("lambertian", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	lambertian := material.Lambertian{
		Tex: texture,
	}
	return lambertian, nil
}

func toMetal(data map[string]interface{}) (material.Material, error) {
	textureData, err := getAttr("metal", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData.(map[string]interface{}))
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
	textureData, err := getAttr("dielectric", data, "texture")
	if err != nil {
		return nil, err
	}
	texture, err := toTexture(textureData.(map[string]interface{}))
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
	emitData, err := getAttr("diffuseLight", data, "emit")
	if err != nil {
		return nil, err
	}
	emit, err := toTexture(emitData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	diffuseLight := material.DiffuseLight{
		Emit: emit,
	}
	return diffuseLight, err
}

func toIsotropic(data map[string]interface{}) (material.Material, error) {
	albedoData, err := getAttr("isotropic", data, "albedo")
	if err != nil {
		return nil, err
	}
	albedo, err := toTexture(albedoData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	isotropic := material.Isotropic{
		Albedo: albedo,
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
	vecData, err := getAttr(t, data, key)
	if err != nil {
		return nil, err
	}
	vec, err := toVec(vecData.(map[string]interface{}))
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
		return 0, errors.New(fmt.Sprintf("%v expected number type of %v: %v", t, key, data))
	}
	return number.(float64), nil
}

func getString(t string, data map[string]interface{}, key string) (string, error) {
	number, err := getAttr(t, data, key)
	if err != nil {
		return "", err
	}
	if reflect.ValueOf(number).Kind() != reflect.String {
		return "", errors.New(fmt.Sprintf("%v expected string type of %v: %v", t, key, data))
	}
	return number.(string), nil
}

func getAttr(t string, data map[string]interface{}, key string) (interface{}, error) {
	attrVal, ok := data[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%v is missing %v in %v", t, key, data))
	}
	return attrVal, nil
}
