package api

import "time"

// createRequest is what a client is allowed to send. It is deliberately
// NOT switches.Switch: State, LastCheckIn, ID and both token are owned
// by the server. Decoding straight into the domain type would let a
// client post {"state": "disarmed"} and refuse its own switch.
type createRequest struct {
	Label     string `json:"label"`
	Secret    string `json:"secret"`
	Recipient string `json:"recipient"`
	Interval  string `json:"interval"` // "72h", "30m" - time.ParseDuration
}

// createResponse is shown exactly once. The tokens never appear again.
type createResponse struct {
	ID           string    `json:"id"`
	CheckInToken string    `json:"check_in_token"`
	RevealToken  string    `json:"reveal_token"`
	Deadline     time.Time `json:"deadline"`
}

// switchResponse is the public view of a switch: no secret, no tokens.
type switchResponse struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	State       string    `json:"state"`
	Interval    string    `json:"interval"`
	LastCheckIn time.Time `json:"last_check_in"`
	Deadline    time.Time `json:"deadline"`
}

// checkInResponse tells the owner what changed. No token, no secret.
type checkInResponse struct {
	State       string    `json:"state"`
	LastCheckIn time.Time `json:"last_check_in"`
	Deadline    time.Time `json:"deadline"`
}
