package cassandra

import (
	"fmt"
	"github.com/satori/go.uuid"
)

func newEmailForTest() string {
	return fmt.Sprintf("brotato-%v@example.com", uuid.NewV4().String())
}
