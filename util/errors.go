package util

import (
	"log"

	"github.com/gin-gonic/gin"
)

// Collate error handling and logging in a single place in case we need/want to
// do something funky (like send logs to a logging server or update a varz or something).
//
// Also, it replaces >1 line of code with 1 line of code.
func DoHttpError(c *gin.Context, code int, err error) {
	if err != nil {
		log.Print(err.Error())
		c.AbortWithError(code, err)
		return
	}

	c.AbortWithStatus(code)
}

func DoError(err error) {
	log.Print(err.Error())
}

func DoDebug(msg string) {
	log.Print(msg)
}
