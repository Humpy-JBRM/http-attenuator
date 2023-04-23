package api

import (
	"fmt"
	config "http-attenuator/facade/config"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PUT /api/v1/config/:name/:value
func SetConfigHandler(c *gin.Context) {
	// Extract the service from the URL
	name := c.Param("name")
	if name == "" {
		err := fmt.Errorf("SetConfigHandler(%s): no config parameter")
		log.Println(err)
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	value := c.Param("value")
	if value == "" {
		err := fmt.Errorf("SetConfigHandler(%s): no config value")
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if boolVal, err := strconv.ParseBool(value); err != nil {
		if intVal, err := strconv.ParseInt(value, 10, 64); err != nil {
			if floatVal, err := strconv.ParseFloat(value, 64); err != nil {
				_ = config.Config().SetString(name, value)
			} else {
				_ = config.Config().SetFloat(name, floatVal)
			}
		} else {
			_ = config.Config().SetInt(name, intVal)
		}
	} else {
		_ = config.Config().SetBool(name, boolVal)
	}

	c.JSON(http.StatusCreated,
		map[string]string{
			name: value,
		},
	)
}
