// Package apis
package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
)

func FetchGpsDevices(baseURL string) ([]internalgps.GpsResponse, error) {
	resp, err := http.Get(baseURL + "/api/v1/gps-devices")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data []internalgps.GpsResponse `json:"data"`
	}
	if err = json.Unmarshal(resBody, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}
