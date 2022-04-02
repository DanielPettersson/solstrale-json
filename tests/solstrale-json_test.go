package tests

import (
	"errors"
	"testing"

	solstralejson "github.com/DanielPettersson/solstrale-json"
	"github.com/stretchr/testify/assert"
)

func TestToSceneEmpty(t *testing.T) {
	scene, err := solstralejson.ToScene([]byte(``))
	assert.Nil(t, scene)
	assert.Equal(t, errors.New("scene is missing world"), err)
}
