package pitcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldIterateOnAllPlaceholdersUpdatingWithSessionValue(t *testing.T) {

	s := DynamicRWSession{
		parameters: map[string]string{
			"with":  "with-value",
			"place": "place-value",
		},
	}
	c := NewClientWithSession(&s)

	txt := "some text ${with} some ${place}holders to be ${with} updated"
	expectedTxt := "some text with-value some place-valueholders to be with-value updated"

	output := c.parseSessionKeys(txt)

	assert.Equal(t, expectedTxt, output)

}

func TestShouldIterateOnAllPlaceholdersDoingNothingOnNotFound(t *testing.T) {

	s := DynamicRWSession{
		parameters: map[string]string{
			"with":  "with-value",
			"place": "place-value",
		},
	}
	c := NewClientWithSession(&s)
	txt := "some text without placeholders to be updated"
	output := c.parseSessionKeys(txt)

	assert.Equal(t, txt, output)

}
