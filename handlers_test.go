package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_application(t *testing.T) {
	pathToConfig := os.Getenv(Config)
	c, err := configuration(pathToConfig)

	if err != nil {
		t.Fatal(err)
	}

	app := application{
		configuration: c,
		errorLog:      log.New(ioutil.Discard, "", 0),
	}

	r := app.router()

	tests := []struct {
		name         string
		sectorId     int
		coords       coordinates
		wantRespBody map[string]interface{}
		wantCode     int
	}{
		{
			name:     "DNS where a drone is located.",
			sectorId: 1,
			coords: coordinates{
				X:   "123.1",
				Y:   "123.2",
				Z:   "123.3",
				Vel: "20.4",
			},
			wantRespBody: map[string]interface{}{
				"loc": float64(390),
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "A direct neighbour DNS.",
			sectorId: 2,
			coords: coordinates{
				X:   "123.1",
				Y:   "123.2",
				Z:   "123.3",
				Vel: "20.4",
			},
			wantRespBody: map[string]interface{}{
				"loc": 759.6,
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "An indirect neighbour DNS.",
			sectorId: 4,
			coords: coordinates{
				X:   "123.1",
				Y:   "123.2",
				Z:   "123.3",
				Vel: "20.4",
			},
			wantRespBody: map[string]interface{}{
				"loc": 1498.8,
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "Undiscovered sector.",
			sectorId: 5,
			coords: coordinates{
				X:   "123.1",
				Y:   "123.2",
				Z:   "123.3",
				Vel: "20.4",
			},
			wantRespBody: map[string]interface{}{
				"message": "Unable to find a direction to the sector 5\n",
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Invalid coordinates.",
			sectorId: 1,
			coords: coordinates{
				X:   "invalidX",
				Y:   "invalidY",
				Z:   "InvalidZ",
				Vel: "InvalidVel",
			},
			wantRespBody: map[string]interface{}{
				"message": "Invalid coordinates",
				"error":   "Key: 'coordinates.X' Error:Field validation for 'X' failed on the 'numeric' tag\nKey: 'coordinates.Y' Error:Field validation for 'Y' failed on the 'numeric' tag\nKey: 'coordinates.Z' Error:Field validation for 'Z' failed on the 'numeric' tag\nKey: 'coordinates.Vel' Error:Field validation for 'Vel' failed on the 'numeric' tag",
			},
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := postRequest(fmt.Sprintf("http://%s%s/%d", app.configuration.Address, SectorDirectionEndpoint, tt.sectorId), &tt.coords)
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			resp := make(map[string]interface{})
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantRespBody, resp)
		})
	}
}

func postRequest(url string, coords *coordinates) (*http.Request, error) {
	data, err := json.Marshal(coords)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
