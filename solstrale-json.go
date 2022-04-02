package solstralejson

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DanielPettersson/solstrale/camera"
	"github.com/DanielPettersson/solstrale/geo"
	"github.com/DanielPettersson/solstrale/hittable"
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
		World:           *world,
		Cam:             *camera,
		BackgroundColor: *background,
	}, nil
}

func toHittable(data map[string]interface{}) (*hittable.Hittable, error) {
	return nil, errors.New("Hittable not implemented")
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
