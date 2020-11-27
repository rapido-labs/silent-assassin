package shifter

import (
	"strconv"
	"time"

	"github.com/stretchr/testify/assert"
)

func (st *ShifterTestSuit) TestIfTimeIsWithinWLInterval() {

	intervalAndResuls := [][]string{
		{"02:00", "03:00", "02:30", "true"},
		{"02:00", "03:00", "04:30", "false"},
		{"23:00", "03:00", "02:30", "true"},
		{"23:00", "03:00", "03:30", "false"},
		{"23:00", "03:00", "22:30", "false"},
	}

	for _, times := range intervalAndResuls {

		start, _ := time.Parse(timeLayout, times[0])
		end, _ := time.Parse(timeLayout, times[1])
		check, _ := time.Parse(timeLayout, times[2])
		result, _ := strconv.ParseBool(times[3])

		assert.Equal(st.T(), timeWithinWLIntervalCheck(start, end, check), result)
	}

}
