package manifest

import (
	"math"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	defaultInitialInterval = 500 * time.Millisecond
	defaultMaxInterval     = 60 * time.Second
	defaultMaxElapsedTime  = 15 * time.Minute
	defaultFactor          = 1.5
	defaultJitter          = 0.5
)

var defaultBackoff = wait.Backoff{
	Duration: defaultInitialInterval,
	Cap:      defaultMaxInterval,
	Steps:    int(math.Ceil(float64(defaultMaxElapsedTime) / float64(defaultInitialInterval))), // now a required argument
	Factor:   defaultFactor,
	Jitter:   defaultJitter,
}
