package checker

import (
	"encoding/json"
	"github.com/soulteary/cppreportcard-action/internal/project"
)

func jsonCommands(c []project.Command) ([]byte, error) { return json.Marshal(c) }
