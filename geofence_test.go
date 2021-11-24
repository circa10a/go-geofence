package geofence

import (
	"errors"
	"testing"
)

func TestFormatCoordinates(t *testing.T) {
	type test struct {
		Expected    string
		Sensitivity int
		Input       float32
	}
	tests := []test{
		{
			Input:       -31.12345,
			Expected:    "-31.12345",
			Sensitivity: 5,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.1234",
			Sensitivity: 4,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.123",
			Sensitivity: 3,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.123",
			Sensitivity: 3,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.12",
			Sensitivity: 2,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.1",
			Sensitivity: 1,
		},
		{
			Input:       -31.12345,
			Expected:    "-31",
			Sensitivity: 0,
		},
	}
	for _, test := range tests {
		actual := formatCoordinates(test.Sensitivity, test.Input)
		if test.Expected != actual {
			t.Fail()
		}
	}
}

func TestValidateSensitivity(t *testing.T) {
	type test struct {
		Expected error
		Input    int
	}
	tests := []test{
		{
			Input:    6,
			Expected: ErrInvalidSensitivity,
		},
		{
			Input:    5,
			Expected: nil,
		},
		{
			Input:    4,
			Expected: nil,
		},
		{
			Input:    3,
			Expected: nil,
		},
		{
			Input:    2,
			Expected: nil,
		},
		{
			Input:    1,
			Expected: nil,
		},
		{
			Input:    0,
			Expected: nil,
		},
		{
			Input:    -1,
			Expected: ErrInvalidSensitivity,
		},
	}
	for _, test := range tests {
		actual := validateSensitivity(test.Input)
		if !errors.Is(actual, test.Expected) {
			t.Fail()
		}
	}
}

func TestValidateIPAddress(t *testing.T) {
	type test struct {
		Expected error
		Input    string
	}
	tests := []test{
		{
			Input:    "8.8.8.8",
			Expected: nil,
		},
		{
			Input:    "8.8.88",
			Expected: ErrInvalidIPAddress,
		},
	}
	for _, test := range tests {
		actual := validateIPAddress(test.Input)
		if !errors.Is(actual, test.Expected) {
			t.Fail()
		}
	}
}
