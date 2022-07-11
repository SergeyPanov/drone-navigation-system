package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

type allyResponse struct {
	code int
	loc  float64
}

func (a *application) logServerError(err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	a.errorLog.Output(2, trace)
}

func (a *application) loc(c *coordinates) (float64, error) {
	bitSize := 64

	xf, err := strconv.ParseFloat(c.X, bitSize)
	if err != nil {
		return -1, err
	}

	yf, err := strconv.ParseFloat(c.Y, bitSize)
	if err != nil {
		return -1, err
	}

	zf, err := strconv.ParseFloat(c.Z, bitSize)
	if err != nil {
		return -1, err
	}

	velf, err := strconv.ParseFloat(c.Vel, bitSize)
	if err != nil {
		return -1, err
	}

	return (xf+yf+zf)*float64(a.configuration.SectorId) + velf, nil
}

func (a *application) requestToAlly(ally string, sectorId int, coords *coordinates) (*http.Request, error) {
	requestBody, err := json.Marshal(coords)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s/%d", ally, SectorDirectionEndpoint, sectorId), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set(AlliesToNotAsk, strings.Join(a.configuration.Allies, ";"))
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (a *application) queryAlly(request *http.Request) (*allyResponse, error) {
	resp, err := a.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var unmarshalled map[string]float64

		err = json.Unmarshal(body, &unmarshalled)
		if err != nil {
			return nil, err
		}

		return &allyResponse{
			code: http.StatusOK,
			loc:  unmarshalled["loc"],
		}, nil
	}

	return &allyResponse{
		code: resp.StatusCode,
		loc:  math.Inf(-1),
	}, nil
}
