package solstralejson

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"

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

	backgroundData, err := getAttr("scene", data, "background")
	if err != nil {
		return nil, err
	}
	background, err := toVec(backgroundData.(map[string]interface{}))
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
	t, err := getAttr("hittable", data, "type")
	if err != nil {
		return nil, err
	}

	switch t.(string) {
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
	for _, item := range list.([]interface{}) {
		hittable, err := toHittable(item.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
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

	density, err := getAttr("constantMedium", data, "density")
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
		density.(float64),
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

	blurDirectionData, err := getAttr("motionBlur", data, "blurDirection")
	if err != nil {
		return nil, err
	}
	blurDirection, err := toVec(blurDirectionData.(map[string]interface{}))
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
	cornerData, err := getAttr("quad", data, "corner")
	if err != nil {
		return nil, err
	}
	corner, err := toVec(cornerData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	dirUData, err := getAttr("quad", data, "dirU")
	if err != nil {
		return nil, err
	}
	dirU, err := toVec(dirUData.(map[string]interface{}))

	dirVData, err := getAttr("quad", data, "dirV")
	if err != nil {
		return nil, err
	}
	dirV, err := toVec(dirVData.(map[string]interface{}))

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

	angle, err := getAttr("rotationY", data, "angle")
	if err != nil {
		return nil, err
	}

	rotationY := hittable.NewRotationY(
		object,
		angle.(float64),
	)

	return rotationY, nil
}

func toSphere(data map[string]interface{}) (hittable.Hittable, error) {
	centerData, err := getAttr("sphere", data, "center")
	if err != nil {
		return nil, err
	}
	center, err := toVec(centerData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	radius, err := getAttr("sphere", data, "radius")
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
		radius.(float64),
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

	offsetData, err := getAttr("translation", data, "offset")
	if err != nil {
		return nil, err
	}
	offset, err := toVec(offsetData.(map[string]interface{}))
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
	t, err := getAttr("texture", data, "type")
	if err != nil {
		return nil, err
	}

	switch t.(string) {
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
	colorData, err := getAttr("solidColor", data, "color")
	if err != nil {
		return nil, err
	}
	color, err := toVec(colorData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	return material.SolidColor{
		ColorValue: *color,
	}, nil
}

func toChecker(data map[string]interface{}) (material.Texture, error) {
	scale, err := getAttr("checker", data, "scale")
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
		Scale: scale.(float64),
		Even:  even,
		Odd:   odd,
	}, nil
}

func toImage(data map[string]interface{}) (material.Texture, error) {
	pathData, err := getAttr("image", data, "path")
	if err != nil {
		return nil, err
	}

	f, err := os.Open(pathData.(string))
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
	colorData, err := getAttr("solidColor", data, "color")
	if err != nil {
		return nil, err
	}
	color, err := toVec(colorData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	scale, err := getAttr("checker", data, "scale")
	if err != nil {
		return nil, err
	}

	noiseTexture := material.NoiseTexture{
		ColorValue: *color,
		Scale:      scale.(float64),
	}

	return noiseTexture, nil
}

func toMaterial(data map[string]interface{}) (material.Material, error) {
	t, err := getAttr("material", data, "type")
	if err != nil {
		return nil, err
	}

	switch t.(string) {
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

	fuzz, err := getAttr("metal", data, "fuzz")
	if err != nil {
		return nil, err
	}

	metal := material.Metal{
		Tex:  texture,
		Fuzz: fuzz.(float64),
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

	indexOfRefraction, err := getAttr("dielectric", data, "indexOfRefraction")
	if err != nil {
		return nil, err
	}

	metal := material.Dielectric{
		Tex:               texture,
		IndexOfRefraction: indexOfRefraction.(float64),
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
	imageWidth, err := getAttr("camera", data, "imageWidth")
	if err != nil {
		return nil, err
	}

	imageHeight, err := getAttr("camera", data, "imageHeight")
	if err != nil {
		return nil, err
	}

	verticalFovDegrees, err := getAttr("camera", data, "verticalFovDegrees")
	if err != nil {
		return nil, err
	}

	aperture, err := getAttr("camera", data, "aperture")
	if err != nil {
		return nil, err
	}

	focusDistance, err := getAttr("camera", data, "focusDistance")
	if err != nil {
		return nil, err
	}

	lookFromData, err := getAttr("camera", data, "lookFrom")
	if err != nil {
		return nil, err
	}
	lookFrom, err := toVec(lookFromData.(map[string]interface{}))

	lookAtData, err := getAttr("camera", data, "lookAt")
	if err != nil {
		return nil, err
	}
	lookAt, err := toVec(lookAtData.(map[string]interface{}))

	vupData, err := getAttr("camera", data, "vup")
	if err != nil {
		return nil, err
	}
	vup, err := toVec(vupData.(map[string]interface{}))

	cam := camera.New(
		int(imageWidth.(float64)),
		int(imageHeight.(float64)),
		verticalFovDegrees.(float64),
		aperture.(float64),
		focusDistance.(float64),
		*lookFrom,
		*lookAt,
		*vup,
	)
	return &cam, nil
}

func toVec(data map[string]interface{}) (*geo.Vec3, error) {
	x, err := getAttr("vec", data, "x")
	if err != nil {
		return nil, err
	}
	y, err := getAttr("vec", data, "y")
	if err != nil {
		return nil, err
	}
	z, err := getAttr("vec", data, "z")
	if err != nil {
		return nil, err
	}
	vec := geo.NewVec3(x.(float64), y.(float64), z.(float64))
	return &vec, nil
}

func getAttr(t string, data map[string]interface{}, key string) (interface{}, error) {
	attrVal, ok := data[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%v is missing %v", t, key))
	}
	return attrVal, nil
}
