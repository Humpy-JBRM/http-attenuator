package data

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PathologyProfileFromConfig map[string]*PathologyImpl

type PathologyProfile interface {
	Handler
	GetPathologyByName(name string) Pathology
	GetPathology() Pathology
}

type PathologyProfileImpl struct {
	name string
	PathologyProfileFromConfig

	// The pathologies in this profile as a CDF
	pathologyCdf []HasCDF
	rng          *rand.Rand
}

func (pp *PathologyProfileImpl) GetName() string {
	return pp.name
}

func (pp *PathologyProfileImpl) GetPathologyByName(name string) Pathology {
	return pp.PathologyProfileFromConfig[name]
}

func (pp *PathologyProfileImpl) GetPathology() Pathology {
	pathology := ChooseFromCDF(pp.rng.Float64(), pp.pathologyCdf)
	if pathology == nil {
		return nil
	}
	return pathology.(Pathology)
}

// Satisfy the Handler duck type
func (pp *PathologyProfileImpl) Handle(c *gin.Context) {
	pathology := pp.GetPathology()
	if pathology == nil {
		err := fmt.Errorf("")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Defer to the chosen pathology
	pathology.(Pathology).Handle(c)
}
