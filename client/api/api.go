package api

import (
	"bytes"
	"dfs/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL string
	key     string
}

func newRequest(method, url string, body io.Reader, key string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func NewClient(baseURL string, key string) *Client {
	return &Client{
		baseURL: baseURL,
		key:     key,
	}
}

func (c *Client) GetObject(name string) (*types.Object, error) {

	var objReq types.GetObjectRequest
	objReq.Name = name

	encoded, err := json.Marshal(objReq)
	if err != nil {
		return nil, err
	}

	req, err := newRequest("POST", c.baseURL+"/object/get", bytes.NewBuffer(encoded), c.key)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var objResp types.GetObjectResponse

	if err := json.NewDecoder(resp.Body).Decode(&objResp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.ErrCouldNotGetObjectFromAPI
	}

	return &objResp.Object, nil
}

func (c *Client) PutObject(obj *types.Object) error {
	encoded, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req, err := newRequest("POST", c.baseURL+"/object/put", bytes.NewBuffer(encoded), c.key)
	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.ErrCouldNotPutObjectToAPI
	}

	return nil
}

func (c *Client) CreateSegment(segment *types.Segment) error {
	encoded, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/objects/%s/segments", segment.ObjectID.String())

	req, err := newRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(encoded), c.key)
	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.ErrCouldNotPutObjectToAPI
	}

	return nil
}
