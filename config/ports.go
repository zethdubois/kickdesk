package config

import (
	"encoding/json"
	"fmt"
)

// CrowdedPorts are common defaults that often collide across unrelated projects.
var CrowdedPorts = map[int]string{
	5173: "Vite default",
	3000: "common frontend default",
	5432: "Docker Postgres default",
}

// PortBinding is one host port with an optional role label.
type PortBinding struct {
	Role string `json:"role"`
	Port int    `json:"port"`
}

// PortsList unmarshals either [5000, 5043] or [{"role":"app","port":5000}, ...].
type PortsList []PortBinding

// UnmarshalJSON accepts a JSON array of integers or port objects.
func (p *PortsList) UnmarshalJSON(data []byte) error {
	var ints []int
	if err := json.Unmarshal(data, &ints); err == nil {
		out := make(PortsList, len(ints))
		for i, n := range ints {
			out[i] = PortBinding{Port: n}
		}
		*p = out
		return nil
	}
	var objs []PortBinding
	if err := json.Unmarshal(data, &objs); err != nil {
		return fmt.Errorf("ports: expected array of integers or {role, port} objects")
	}
	*p = objs
	return nil
}

// Numbers returns host port numbers in config order.
func (p PortsList) Numbers() []int {
	out := make([]int, len(p))
	for i, b := range p {
		out[i] = b.Port
	}
	return out
}

// FamilyLabel returns a glance label like "50xx" for port_family 50.
func FamilyLabel(family int) string {
	if family <= 0 {
		return ""
	}
	return fmt.Sprintf("%dxx", family)
}
