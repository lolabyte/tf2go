package gen_test

import (
	"testing"

	"github.com/lolabyte/tf2go/gen"
	"github.com/stretchr/testify/assert"
)

func Test__generateTFmodulePackage(t *testing.T) {
	t.Run("returns an error when the input module directory doesn't exist", func(t *testing.T) {
		err := gen.GenerateTFModulePackage("does_not_exist", "out_dir", "test_module")
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "Failed to read module directory: Module directory does_not_exist does not exist or cannot be read.")
	})

	t.Run("returns an error when input module is invalid", func(t *testing.T) {
		err := gen.GenerateTFModulePackage("../testdata/invalid_tf_module", "out_dir", "test_module")
		assert.Error(t, err)
		assert.Regexp(t, "Argument or block definition required:.*$", err.Error())
	})
}
