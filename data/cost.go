package data

// CostFromConfig is what gets read from config.
// Key is the name(s) of the 'coins'
type CostFromConfig map[string]int

type Cost interface {
	Charge() (bool, error)
}

type CostImpl struct {
	coins CostFromConfig
}

func NewCost(coins CostFromConfig) Cost {
	return &CostImpl{
		coins: coins,
	}
}

func (c *CostImpl) Charge() (bool, error) {
	return false, nil
}
