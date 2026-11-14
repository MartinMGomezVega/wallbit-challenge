package y

import (
	"time"

	_ "fake/internal/interfaces/http"
)

func Now() time.Time { return time.Now() }
