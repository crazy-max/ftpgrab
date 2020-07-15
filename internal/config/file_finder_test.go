package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinderFind(t *testing.T) {
	configFile, err := ioutil.TempFile("", "ftpgrab-file-finder-test-*.yml")
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(configFile.Name())
	}()

	dir, err := ioutil.TempDir("", "ftpgrab-file-finder-test")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	fooFile, err := os.Create(filepath.Join(dir, "foo.yml"))
	require.NoError(t, err)

	_, err = os.Create(filepath.Join(dir, "bar.yml"))
	require.NoError(t, err)

	type expected struct {
		error bool
		path  string
	}

	testCases := []struct {
		desc       string
		basePaths  []string
		configFile string
		expected   expected
	}{
		{
			desc:       "not found: no config file",
			configFile: "",
			expected:   expected{path: ""},
		},
		{
			desc:       "not found: no config file, no other paths available",
			configFile: "",
			basePaths:  []string{"/my/path/ftpgrab", "$HOME/my/path/ftpgrab", "./my-ftpgrab"},
			expected:   expected{path: ""},
		},
		{
			desc:       "not found: with non existing config file",
			configFile: "/my/path/config.yml",
			expected:   expected{path: ""},
		},
		{
			desc:       "found: with config file",
			configFile: configFile.Name(),
			expected:   expected{path: configFile.Name()},
		},
		{
			desc:       "found: no config file, first base path",
			configFile: "",
			basePaths:  []string{filepath.Join(dir, "foo"), filepath.Join(dir, "bar")},
			expected:   expected{path: fooFile.Name()},
		},
		{
			desc:       "found: no config file, base path",
			configFile: "",
			basePaths:  []string{"/my/path/ftpgrab", "$HOME/my/path/ftpgrab", filepath.Join(dir, "foo")},
			expected:   expected{path: fooFile.Name()},
		},
		{
			desc:       "found: config file over base path",
			configFile: configFile.Name(),
			basePaths:  []string{filepath.Join(dir, "foo"), filepath.Join(dir, "bar")},
			expected:   expected{path: configFile.Name()},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			finder := Finder{
				BasePaths:  test.basePaths,
				Extensions: []string{"yaml", "yml"},
			}

			path, err := finder.Find(test.configFile)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.path, path)
			}
		})
	}
}

func TestFinderGetPaths(t *testing.T) {
	testCases := []struct {
		desc       string
		basePaths  []string
		configFile string
		expected   []string
	}{
		{
			desc:       "no config file",
			basePaths:  []string{"/etc/ftpgrab/ftpgrab", "$HOME/.config/ftpgrab", "./ftpgrab"},
			configFile: "",
			expected: []string{
				"/etc/ftpgrab/ftpgrab.yaml",
				"/etc/ftpgrab/ftpgrab.yml",
				"$HOME/.config/ftpgrab.yaml",
				"$HOME/.config/ftpgrab.yml",
				"./ftpgrab.yaml",
				"./ftpgrab.yml",
			},
		},
		{
			desc:       "with config file",
			basePaths:  []string{"/etc/ftpgrab/ftpgrab", "$HOME/.config/ftpgrab", "./ftpgrab"},
			configFile: "/my/path/config.yml",
			expected: []string{
				"/my/path/config.yml",
				"/etc/ftpgrab/ftpgrab.yaml",
				"/etc/ftpgrab/ftpgrab.yml",
				"$HOME/.config/ftpgrab.yaml",
				"$HOME/.config/ftpgrab.yml",
				"./ftpgrab.yaml",
				"./ftpgrab.yml",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			finder := Finder{
				BasePaths:  test.basePaths,
				Extensions: []string{"yaml", "yml"},
			}
			paths := finder.getPaths(test.configFile)

			assert.Equal(t, test.expected, paths)
		})
	}
}
