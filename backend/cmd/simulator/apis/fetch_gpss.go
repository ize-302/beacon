// Package apis
package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
)

func APIFetchGpss(baseURL string) ([]internalgps.GpsResponse, error) {
	resp, err := http.Get(baseURL + "/gps")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	gpss := []internalgps.GpsResponse{}

	if err = json.Unmarshal(resBody, &gpss); err != nil {
		return nil, err
	}
	return gpss, nil
}
