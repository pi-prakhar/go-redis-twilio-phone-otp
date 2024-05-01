package api

type OTPData struct {
	PhoneNumber string `json:"phoneNumber,omitempty" validate:"required"`
}

type VerifyData struct {
	User *OTPData `json:"user,omitempty" validate:"required"`
	Code string   `json:"code,omitempty" validate:"required"`
}

type TimeData struct {
	User *OTPData `json:"user,omitempty" validate:"required"`
	TTL  int      `json:"ttl,omitempty" validate:"required"`
}
type TrialsLeft struct {
	User   *OTPData `json:"user,omitempty" validate:"required"`
	Trials int      `json:"trials,omitempty" validate:"required"`
}
