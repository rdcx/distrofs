package api_test

import (
	"dfs/client/api"
	"dfs/types"
	"testing"

	"github.com/h2non/gock"
)

func TestGetObject(t *testing.T) {
	t.Run("can get object", func(t *testing.T) {
		defer gock.Off()

		mockRes := types.GetObjectResponse{
			Object: types.NewObject("/home/john/file.txt"),
		}

		gock.New("http://localhost:8080").
			JSON(`{"name":"/home/john/file.txt"}`).
			Post("/object/get").
			Reply(200).
			JSON(mockRes)

		api := api.NewClient("http://localhost:8080", "123")
		obj := types.Object{
			ID:   mockRes.Object.ID,
			Name: mockRes.Object.Name,
		}
		expected := types.GetObjectResponse{
			Object: obj,
		}

		actual, err := api.GetObject(obj.Name)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if actual.ID != expected.Object.ID {
			t.Errorf("expected %v, got %v", expected.Object.ID, actual.ID)
		}

		if actual.Name != expected.Object.Name {
			t.Errorf("expected %v, got %v", expected.Object.Name, actual.Name)
		}
	})

	t.Run("can handle not found", func(t *testing.T) {
		defer gock.Off()

		gock.New("http://localhost:8080").
			JSON(`{"name":"/home/john/file.txt"}`).
			Post("/object/get").
			Reply(404).
			JSON(`{"error":"not found"}`)

		api := api.NewClient("http://localhost:8080", "123")
		obj := types.NewObject("/home/john/file.txt")

		_, err := api.GetObject(obj.Name)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("can handle error", func(t *testing.T) {
		defer gock.Off()

		gock.New("http://localhost:8080").
			JSON(`{"name":"/home/john/file.txt"}`).
			Post("/object/get").
			Reply(500).
			JSON(`{"error":"internal server error"}`)

		api := api.NewClient("http://localhost:8080", "123")
		obj := types.NewObject("/home/john/file.txt")

		_, err := api.GetObject(obj.Name)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestPutObject(t *testing.T) {
	t.Run("can put object", func(t *testing.T) {
		defer gock.Off()

		mockObj := types.NewObject("/home/john/file.txt")

		gock.New("http://localhost:8080").
			JSON(mockObj).
			Post("/object/put").
			Reply(200).
			JSON(`{"success":"ok"}`)

		api := api.NewClient("http://localhost:8080", "123")

		err := api.PutObject(&mockObj)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("can handle error", func(t *testing.T) {
		defer gock.Off()

		mockObj := types.NewObject("/home/john/file.txt")

		gock.New("http://localhost:8080").
			JSON(mockObj).
			Post("/object/put").
			Reply(500).
			JSON(`{"error":"internal server error"}`)

		api := api.NewClient("http://localhost:8080", "123")

		err := api.PutObject(&mockObj)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
