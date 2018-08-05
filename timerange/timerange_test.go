package timerange

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Apply(t *testing.T) {
	ref := time.Now().UTC()

	actual, err := Apply(ref, "-1h")
	assert.Nil(t, err)
	assert.Equal(t, ref.Add(-1*time.Hour), actual)

	actual, err = Apply(ref, "2mo")
	assert.Nil(t, err)
	assert.Equal(t, ref.AddDate(0, 2, 0), actual)

	actual, err = Apply(ref, "-1y 3h 126s")
	assert.Nil(t, err)
	assert.Equal(t, ref.AddDate(-1, 0, 0).Add(-3*time.Hour-126*time.Second), actual)

	actual, err = Apply(ref, "3d 1h 6m")
	assert.Nil(t, err)
	assert.Equal(t, ref.AddDate(0, 0, 3).Add(time.Hour+6*time.Minute), actual)
}

func Test_Apply_Fail(t *testing.T) {
	ref := time.Now().UTC()

	actual, err := Apply(ref, "42")
	assert.Equal(t, ErrInvalidRange, err)
	assert.Equal(t, ref, actual)
}

func Test_FromDuration(t *testing.T) {
	ref := time.Now().UTC()

	actual := FromDuration(ref.Sub(ref.Add(1 * time.Hour)))
	assert.Equal(t, "-1h", actual)

	actual = FromDuration(-1 * ref.Sub(ref.AddDate(0, 0, 60)))
	assert.Equal(t, "60d", actual)

	actual = FromDuration(ref.Sub(ref.AddDate(0, 0, 1).Add(3*time.Hour + 126*time.Second)))
	assert.Equal(t, "-1d 3h 2m 6s", actual)

	actual = FromDuration(-1 * ref.Sub(ref.AddDate(0, 0, 3).Add(time.Hour+6*time.Minute)))
	assert.Equal(t, "3d 1h 6m", actual)
}
