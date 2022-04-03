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
		"scene-root-missing-world.json": errors.New("scene is missing world in map[a:b]"),
	}

	for fileName, expectedErr := range testCases {

		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		scene, err := solstralejson.ToScene(b)
		assert.Nil(t, scene)
		assert.Equal(t, expectedErr, err)
	}
}
