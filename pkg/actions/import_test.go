// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package actions

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/stretchr/testify/require"
)

func TestImport_http(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		f, err := os.Open(dataPath)
		require.NoError(t, err)

		defer f.Close()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Disposition", `attachment; filename="manifest.yaml"`)
			http.ServeContent(w, r, "file.yaml", time.Time{}, f)
		}))
		defer ts.Close()

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: "/",
			OptionPath:   ts.URL,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Contains(t, name, "service-my-service-")
			assert.Equal(t, "", moduleName)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)

	})
}

func TestImport_yaml_file(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		module := "/"
		path := "/file.yaml"

		stageFile(t, appMock.Fs(), "import/file.yaml", path)

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Contains(t, name, "service-my-service-")
			assert.Equal(t, "", moduleName)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestImport_json_file(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "my-service.json")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		path := "/my-service.json"

		stageFile(t, appMock.Fs(), "import/my-service.json", path)

		cases := []struct {
			name           string
			module         string
			path           string
			expectedModule string
		}{
			{
				name:           "root module",
				module:         "/",
				expectedModule: "/",
			},
			{
				name:           "module",
				module:         "a",
				expectedModule: "a",
			},
			{
				name:           "dot module",
				module:         "a.b",
				expectedModule: "a.b",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				in := map[string]interface{}{
					OptionApp:    appMock,
					OptionModule: tc.module,
					OptionPath:   path,
				}

				a, err := NewImport(in)
				require.NoError(t, err)

				a.createComponentFn = func(_ app.App, moduleName, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
					assert.Contains(t, name, "my-service")
					assert.Equal(t, tc.expectedModule, moduleName)
					assert.Equal(t, string(serviceData), text)
					assert.Equal(t, params.Params{}, p)
					assert.Equal(t, prototype.JSON, templateType)

					return "/", nil
				}

				err = a.Run()
				require.NoError(t, err)
			})
		}
	})
}

func TestImport_directory(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		module := "/"
		path := "/import"

		stageFile(t, appMock.Fs(), "import/file.yaml", "/import/file.yaml")

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Contains(t, name, "service-my-service-")
			assert.Equal(t, "", moduleName)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestImport_invalid_file(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		module := "/"
		path := "/import"

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			return "", errors.New("invalid")
		}

		err = a.Run()
		require.Error(t, err)
	})
}

func TestImport_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewImport(in)
	require.Error(t, err)
}
