package apis

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
)

func SubscribeToNewDevices(ctx context.Context, baseURL string, onNew func(internalgps.GpsResponse)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/gps-devices/events", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var gps internalgps.GpsResponse
		if err := json.Unmarshal([]byte(data), &gps); err != nil {
			continue
		}
		onNew(gps)
	}
	return scanner.Err()
}
