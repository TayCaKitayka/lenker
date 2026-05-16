package storage

import (
	"strings"
	"testing"
)

func TestCreateSubscriptionSQLCastsStartParameter(t *testing.T) {
	if !strings.Contains(createSubscriptionSQL, "$3::timestamptz") {
		t.Fatalf("create subscription SQL must cast starts_at parameter: %s", createSubscriptionSQL)
	}
	if strings.Contains(createSubscriptionSQL, " $3, $3 + ") {
		t.Fatalf("create subscription SQL must not reuse untyped $3 in interval expression: %s", createSubscriptionSQL)
	}
}
