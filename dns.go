package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const AlliesToNotAsk = "Allies-To-Not-Ask"
const Config = "DNS_CONFIG"
const SectorDirectionEndpoint = "/sector/direction"

type application struct {
	http.Client
	configuration *conf
	externalIp    *string
	errorLog      *log.Logger
}

type coordinates struct {
	X   string `json:"x" binding:"required,numeric"`
	Y   string `json:"y" binding:"required,numeric"`
	Z   string `json:"z" binding:"required,numeric"`
	Vel string `json:"vel" binding:"required,numeric"`
}

type conf struct {
	SectorId int      `yaml:"sectorId"`
	Address  string   `yaml:"address"`
	Allies   []string `yaml:"allies"`
}

func main() {
	errLog := log.New(os.Stdout, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	config := os.Getenv(Config)

	if len(config) <= 0 {
		config = "./dns.yaml"
	}

	yamlConfig, err := ioutil.ReadFile(config)

	if err != nil {
		errLog.Fatal("Unable to process configuration", err)
	}

	c := &conf{}
	err = yaml.Unmarshal(yamlConfig, c)

	if err != nil {
		errLog.Fatal("Corrupted config file", err)
	}

	app := application{
		configuration: c,
		errorLog:      errLog,
	}

	r := gin.Default()
	r.POST(fmt.Sprintf("%s/:sectorId", SectorDirectionEndpoint), app.directionToSector)
	err = r.Run(app.configuration.Address)

	if err != nil {
		app.errorLog.Fatal("Unable to start server ", err)
	}
}

func (a *application) directionToSector(c *gin.Context) {
	secId := c.Param("sectorId")
	id, err := strconv.Atoi(secId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid sectorId",
			"error":   err,
		})
		return
	}

	coords := &coordinates{}
	err = c.ShouldBindJSON(coords)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid coordinates",
			"error":   err,
		})
		return
	}

	if id == a.configuration.SectorId {
		loc, err := a.loc(coords)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Unable to parse one of the coordinates",
				"error":   err,
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
			"message": fmt.Sprintf("Unnable to find a direction to the sector %d\n", id),
		})

		return
	}
}
