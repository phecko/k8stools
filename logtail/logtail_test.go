package logtail

import (
	"sort"
	"testing"
)

func TestLogLinesSort(t *testing.T) {


	unsortLine := LogLines{
		&LogLine{
			Time: LogTimestamp("2019-04-08T12:25:45.321635324Z"),
		},
		&LogLine{
			Time: LogTimestamp("2019-04-08T12:30:09.336397444Z"),
		},
		&LogLine{
			Time: LogTimestamp("2019-04-08T12:24:09.336397444Z"),
		},
	}


	sort.Sort(unsortLine)

	if string(unsortLine[0].Time) != "2019-04-08T12:24:09.336397444Z" {
		t.Fail()
	}

}