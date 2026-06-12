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
	resp, err := http.Get(baseURL + "/gps-devices")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	gpsDevices := []internalgps.GpsResponse{}

	if err = json.Unmarshal(resBody, &gpsDevices); err != nil {
		return nil, err
	}
	return gpsDevices, nil
}
