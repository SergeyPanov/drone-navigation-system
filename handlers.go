package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func (a *application) directionToSector(c *gin.Context) {
	secId := c.Param("sectorId")
	id, err := strconv.Atoi(secId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid sectorId",
			"error":   err.Error(),
		})
		return
	}

	coords := &coordinates{}
	err = c.ShouldBindJSON(coords)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid coordinates",
			"error":   err.Error(),
		})
		return
	}

	if id == a.configuration.SectorId {
		loc, err := a.loc(coords)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to parse one of the coordinates",
				"error":   err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"loc": loc,
		})

		return
	} else {
		alliesToNotAsk := c.GetHeader(AlliesToNotAsk)

		for _, ally := range a.configuration.Allies {
			if strings.Contains(alliesToNotAsk, ally) {
				continue
			}

			request, err := a.requestToAlly(ally, id, coords)
			if err != nil {
				a.logServerError(err)
				continue
			}

			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}

			allies := request.Header.Get(AlliesToNotAsk)
			request.Header.Set(AlliesToNotAsk, strings.Join([]string{allies, fmt.Sprintf("%s://%s", scheme, a.configuration.Address)}, ";"))

			allyResp, err := a.queryAlly(request)
			if err != nil {
				a.logServerError(err)
				continue
			}

			if allyResp.code == http.StatusOK {
				c.JSON(http.StatusOK, gin.H{
					"loc": allyResp.loc,
				})

				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Unable to find a direction to the sector %d\n", id),
		})

		return
	}
}
