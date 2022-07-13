package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const AlliesToNotAsk = "Allies-To-Not-Ask"
const Config = "DNS_CONFIG"
const SectorDirectionEndpoint = "/sector/direction"

type application struct {
	http.Client
	configuration *config
	errorLog      *log.Logger
}

type coordinates struct {
	X   string `json:"x" binding:"required,numeric"`
	Y   string `json:"y" binding:"required,numeric"`
	Z   string `json:"z" binding:"required,numeric"`
	Vel string `json:"vel" binding:"required,numeric"`
}

type config struct {
	SectorId int      `yaml:"sectorId"`
	Address  string   `yaml:"address"`
	Allies   []string `yaml:"allies"`
}

func main() {
	errLog := log.New(os.Stdout, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	pathToConfig := os.Getenv(Config)

	c, err := configuration(pathToConfig)
	if err != nil {
		errLog.Fatal(err)
	}

	app := application{
		configuration: c,
		errorLog:      errLog,
	}

	router := app.router()
	err = router.Run(app.configuration.Address)
	if err != nil {
		app.errorLog.Fatal("Unable to start server ", err)
	}
}

func configuration(path string) (*config, error) {
	yamlConfig, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to process configuration from %s. Check if %s environment variable was set correctly.\n%s", path, Config, err.Error()))
	}

	c := &config{}
	err = yaml.Unmarshal(yamlConfig, c)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to unmarshal the config file %s.\n%s", Config, err.Error()))
	}

	return c, nil
}

func (a *application) router() *gin.Engine {
	r := gin.Default()
	r.POST(fmt.Sprintf("%s/:sectorId", SectorDirectionEndpoint), a.directionToSector)

	return r
}
