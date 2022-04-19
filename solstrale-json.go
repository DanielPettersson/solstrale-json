// Package solstralejson provides functions to convert json data to a solstrale scene object
package solstralejson

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
	"github.com/DanielPettersson/solstrale/material"
	"github.com/DanielPettersson/solstrale/post"
	"github.com/DanielPettersson/solstrale/renderer"
	"github.com/qri-io/jsonschema"
)

var (
	//go:embed schema.json
	schemaBytes []byte
	schema      *jsonschema.Schema = &jsonschema.Schema{}
	schemCtx    context.Context    = context.Background()

	imageCache map[string]image.Image = make(map[string]image.Image)
)

func init() {
	if err := json.Unmarshal(schemaBytes, schema); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}

// ToScene takes a slice of bytes representing json as input and returns a scene.
// If json is not properly formatted an error is returned describing the formatting issue.
func ToScene(jsonBytes []byte) (*renderer.Scene, error) {

	err := validateSchema(jsonBytes)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	json.Unmarshal(jsonBytes, &data)

	return toScene(data)
}

func validateSchema(jsonBytes []byte) error {

	context.Background()
	result, err := schema.ValidateBytes(schemCtx, jsonBytes)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return nil
	} else {
		var msgs []string
		for _, e := range result {
			msgs = append(msgs, e.Message)
		}
		return errors.New(strings.Join(msgs, "\n"))
	}
}

func toScene(data map[string]interface{}) (*renderer.Scene, error) {
	world, err := getHittable(data, "world")
	if err != nil {
		return nil, err
	}

	renderConfig := toRenderConfig(getObject(data, "renderConfig"))
	camera := toCamera(getObject(data, "camera"), renderConfig.ImageWidth, renderConfig.ImageHeight)

	return &renderer.Scene{
		World:           world,
		Cam:             camera,
		BackgroundColor: getColor(data, "background"),
		RenderConfig:    renderConfig,
	}, nil
}

func toRenderConfig(data map[string]interface{}) renderer.RenderConfig {
	return renderer.RenderConfig{
		ImageWidth:      getInt(data, "imageWidth"),
		ImageHeight:     getInt(data, "imageHeight"),
		SamplesPerPixel: getInt(data, "samplesPerPixel"),
		Shader:          toShader(getObject(data, "shader")),
		PostProcessor:   toPostProcessor(getNillableObject(data, "postProcessor")),
	}
}

func toShader(data map[string]interface{}) renderer.Shader {
	switch getString(data, "type") {
	case "pathTracing":
		return toPathTracing(data)
	case "albedo":
		return renderer.AlbedoShader{}
	case "normal":
		return renderer.NormalShader{}
	default: // simple
		return renderer.SimpleShader{}
	}
}

func toPathTracing(data map[string]interface{}) renderer.Shader {
	return renderer.PathTracingShader{
		MaxDepth: getInt(data, "maxDepth"),
	}
}

func toPostProcessor(data map[string]interface{}) post.PostProcessor {
	if data == nil {
		return nil
	}

	switch getString(data, "type") {
	default: // oidn
		return toOidn(data)
	}
}

func toOidn(data map[string]interface{}) post.PostProcessor {
	return post.OidnPostProcessor{
		OidnDenoiseExecutablePath: getString(data, "oidnDenoiseExecutablePath"),
	}
}

func getHittable(data map[string]interface{}, key string) (hittable.Hittable, error) {
	object := getObject(data, key)
	return toHittable(object)
}

func toHittable(data map[string]interface{}) (hittable.Hittable, error) {
	switch getString(data, "type") {
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
	case "box":
		return toBox(data)
	case "rotationY":
		return toRotationY(data)
	case "sphere":
		return toSphere(data)
	case "triangle":
		return toTriangle(data)
	case "objModel":
		return toObjModel(data)
	default: // translation
		return toTranslation(data)
	}
}

func toBvh(data map[string]interface{}) (hittable.Hittable, error) {
	list := data["list"].([]interface{})

	items := make([]hittable.Hittable, 0, len(list))
	for _, itemData := range list {
		item := toObject(itemData)
		hittable, err := toHittable(item)
		if err != nil {
			return nil, err
		}
		items = append(items, hittable)
	}

	bvh := hittable.NewBoundingVolumeHierarchy(items)
	return bvh, nil
}

func toConstantMedium(data map[string]interface{}) (hittable.Hittable, error) {
	object, err := getHittable(data, "object")
	if err != nil {
		return nil, err
	}
	texture, err := getTexture(data, "texture")
	if err != nil {
		return nil, err
	}

	constantMedium := hittable.NewConstantMedium(
		object,
		getFloat(data, "density"),
		texture,
	)

	return constantMedium, nil
}

func toHittableList(data map[string]interface{}) (hittable.Hittable, error) {
	list := data["list"].([]interface{})

	items := hittable.NewHittableList()
	for _, itemData := range list {
		item := toObject(itemData)
		hittable, err := toHittable(item)
		if err != nil {
			return nil, err
		}
		items.Add(hittable)
	}

	return &items, nil
}

func toMotionBlur(data map[string]interface{}) (hittable.Hittable, error) {
	object, err := getHittable(data, "object")
	if err != nil {
		return nil, err
	}

	motionBlur := hittable.NewMotionBlur(
		object,
		getVec(data, "blurDirection"),
	)

	return motionBlur, nil
}

func toQuad(data map[string]interface{}) (hittable.Hittable, error) {
	mat, err := getMaterial(data, "mat")
	if err != nil {
		return nil, err
	}

	quad := hittable.NewQuad(
		getVec(data, "corner"),
		getVec(data, "dirU"),
		getVec(data, "dirV"),
		mat,
	)

	return quad, nil
}

func toBox(data map[string]interface{}) (hittable.Hittable, error) {
	mat, err := getMaterial(data, "mat")
	if err != nil {
		return nil, err
	}

	box := hittable.NewBox(
		getVec(data, "corner"),
		getVec(data, "diagonalCorner"),
		mat,
	)

	return box, nil
}

func toRotationY(data map[string]interface{}) (hittable.Hittable, error) {
	object, err := getHittable(data, "object")
	if err != nil {
		return nil, err
	}

	rotationY := hittable.NewRotationY(
		object,
		getFloat(data, "angle"),
	)

	return rotationY, nil
}

func toSphere(data map[string]interface{}) (hittable.Hittable, error) {
	mat, err := getMaterial(data, "mat")
	if err != nil {
		return nil, err
	}

	sphere := hittable.NewSphere(
		getVec(data, "center"),
		getFloat(data, "radius"),
		mat,
	)

	return sphere, nil
}

func toTriangle(data map[string]interface{}) (hittable.Hittable, error) {
	mat, err := getMaterial(data, "mat")
	if err != nil {
		return nil, err
	}

	triangle := hittable.NewTriangle(
		getVec(data, "v0"),
		getVec(data, "v1"),
		getVec(data, "v2"),
		mat,
	)

	return triangle, nil
}

func toObjModel(data map[string]interface{}) (hittable.Hittable, error) {

	matData := getNillableObject(data, "mat")
	path := getString(data, "path")

	if matData != nil {
		mat, err := toMaterial(matData)
		if err != nil {
			return nil, err
		}

		objModel, err := hittable.NewObjModelWithDefaultMaterial(path, mat)
		if err != nil {
			return nil, err
		}

		return objModel, nil
	}

	objModel, err := hittable.NewObjModel(path)
	if err != nil {
		return nil, err
	}

	return objModel, nil
}

func toTranslation(data map[string]interface{}) (hittable.Hittable, error) {
	object, err := getHittable(data, "object")
	if err != nil {
		return nil, err
	}

	translation := hittable.NewTranslation(
		object,
		getVec(data, "offset"),
	)

	return translation, nil
}

func getTexture(data map[string]interface{}, key string) (material.Texture, error) {
	object := getObject(data, key)
	return toTexture(object)
}

func toTexture(data map[string]interface{}) (material.Texture, error) {
	t := getString(data, "type")
	switch t {
	case "solidColor":
		return toSolidColor(data), nil
	case "checker":
		return toChecker(data)
	case "image":
		return toImage(data)
	default: // noise
		return toNoise(data), nil
	}
}

func toSolidColor(data map[string]interface{}) material.Texture {
	color := getColor(data, "color")
	return material.SolidColor{
		ColorValue: color,
	}
}

func toChecker(data map[string]interface{}) (material.Texture, error) {
	even, err := getTexture(data, "even")
	if err != nil {
		return nil, err
	}
	odd, err := getTexture(data, "odd")
	if err != nil {
		return nil, err
	}
	return material.CheckerTexture{
		Scale: getFloat(data, "scale"),
		Even:  even,
		Odd:   odd,
	}, nil
}

func toImage(data map[string]interface{}) (material.Texture, error) {
	path := getString(data, "path")

	var im image.Image

	cachedImage, cached := imageCache[path]
	if cached {
		im = cachedImage
	} else {

		f, err := os.Open(path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed loading image: %v", err.Error()))
		}
		defer f.Close()
		im, _, err = image.Decode(f)
		if err != nil {
			return nil, err
		}
		imageCache[path] = im
	}

	imageTexture := material.NewImageTexture(im, getBool(data, "mirror"))
	return imageTexture, nil
}

func toNoise(data map[string]interface{}) material.Texture {
	return material.NoiseTexture{
		ColorValue: getColor(data, "color"),
		Scale:      getFloat(data, "scale"),
	}
}

func getMaterial(data map[string]interface{}, key string) (material.Material, error) {
	object := getObject(data, key)
	return toMaterial(object)
}

func toMaterial(data map[string]interface{}) (material.Material, error) {
	switch getString(data, "type") {
	case "lambertian":
		return toLambertian(data)
	case "metal":
		return toMetal(data)
	case "dielectric":
		return toDielectric(data)
	default: // "diffuseLight"
		return toDiffuseLight(data), nil
	}
}

func toLambertian(data map[string]interface{}) (material.Material, error) {
	texture, err := getTexture(data, "texture")
	if err != nil {
		return nil, err
	}

	lambertian := material.Lambertian{
		Tex: texture,
	}
	return lambertian, nil
}

func toMetal(data map[string]interface{}) (material.Material, error) {
	texture, err := getTexture(data, "texture")
	if err != nil {
		return nil, err
	}

	metal := material.Metal{
		Tex:  texture,
		Fuzz: getFloat(data, "fuzz"),
	}
	return metal, nil
}

func toDielectric(data map[string]interface{}) (material.Material, error) {
	texture, err := getTexture(data, "texture")
	if err != nil {
		return nil, err
	}

	metal := material.Dielectric{
		Tex:               texture,
		IndexOfRefraction: getFloat(data, "indexOfRefraction"),
	}
	return metal, nil
}

func toDiffuseLight(data map[string]interface{}) material.Material {
	return material.DiffuseLight{
		Emit: material.SolidColor{
			ColorValue: getColor(data, "color"),
		},
	}
}

func toCamera(data map[string]interface{}, imageWidth, imageHeight int) camera.Camera {
	verticalFovDegrees := getFloat(data, "verticalFovDegrees")
	apertureSize := getFloat(data, "apertureSize")
	focusDistance := getFloat(data, "focusDistance")
	lookFrom := getVec(data, "lookFrom")
	lookAt := getVec(data, "lookAt")
	vup := getVec(data, "vup")

	return camera.New(
		imageWidth,
		imageHeight,
		verticalFovDegrees,
		apertureSize,
		focusDistance,
		lookFrom,
		lookAt,
		vup,
	)
}

func getVec(data map[string]interface{}, key string) geo.Vec3 {
	vecData := getObject(data, key)
	return toVec(vecData)
}

func toVec(data map[string]interface{}) geo.Vec3 {
	x := getFloat(data, "x")
	y := getFloat(data, "y")
	z := getFloat(data, "z")
	return geo.NewVec3(x, y, z)
}

func getColor(data map[string]interface{}, key string) geo.Vec3 {
	colorData := getObject(data, key)
	return toColor(colorData)
}

func toColor(data map[string]interface{}) geo.Vec3 {
	r := getFloat(data, "r")
	g := getFloat(data, "g")
	b := getFloat(data, "b")
	return geo.NewVec3(r, g, b)
}

func getFloat(data map[string]interface{}, key string) float64 {
	return data[key].(float64)
}

func getInt(data map[string]interface{}, key string) int {
	return int(data[key].(float64))
}

func getString(data map[string]interface{}, key string) string {
	return data[key].(string)
}

func getBool(data map[string]interface{}, key string) bool {
	return data[key].(bool)
}

func getObject(data map[string]interface{}, key string) map[string]interface{} {
	return toObject(data[key])
}

func toObject(data interface{}) map[string]interface{} {
	return data.(map[string]interface{})
}

func getNillableObject(data map[string]interface{}, key string) map[string]interface{} {
	return toNillableObject(data[key])
}

func toNillableObject(data interface{}) map[string]interface{} {
	if data == nil {
		return nil
	} else {
		return data.(map[string]interface{})
	}
}
