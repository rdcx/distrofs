package fs

import (
	"dfs/client/api"
	"dfs/network"
	"dfs/progress"
	"dfs/types"
	"io"
)

type FS struct {
	apiClient *api.Client
	network   *network.NodeNetwork
}

func NewFS(baseURL, apiKey string) *FS {
	apiClient := api.NewClient(baseURL, apiKey)
	return &FS{
		apiClient: apiClient,
		network:   network.NewNodeNetwork(apiClient),
	}
}

func (fs *FS) ReadFile(name string, w io.WriteCloser, pc progress.Callback) (*types.Object, error) {
	obj, err := fs.apiClient.GetObject(name)
	if err != nil {
		return nil, err
	}

	fs.network.ReadObject(obj, w, pc)

	return obj, nil
}

func (fs *FS) WriteFile(name string, r io.ReadCloser, pc progress.Callback) (*types.Object, error) {
	obj := types.NewObject(name)
	err := fs.apiClient.PutObject(&obj)
	if err != nil {
		return nil, err
	}

	fs.network.WriteData(obj, r, pc)

	return obj, nil
}
