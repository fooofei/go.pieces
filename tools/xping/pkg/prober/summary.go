package prober

import (
	"bytes"
	"fmt"
	"github.com/montanaflynn/stats"
	"time"
)

// summary will generate a summary for given sent/received and time cost
func summary(sent int64, received int64, durs []time.Duration) string {
	w := &bytes.Buffer{}
	statsData := stats.LoadRawData(durs)

	fmt.Fprintf(w, " Sent = %v, ", sent)
	fmt.Fprintf(w, " Received = %v ", received)
	fmt.Fprintf(w, "(%.1f%s)\n",
		float64(received*100)/float64(sent), "%%")

	min, _ := stats.Min(statsData)
	fmt.Fprintf(w, " Minimum = %.2f ms,", min/1000/1000)
	max, _ := stats.Max(statsData)
	fmt.Fprintf(w, " Maximum = %.2f ms\n", max/1000/1000)
	ave, _ := stats.Mean(statsData)
	fmt.Fprintf(w, " Average = %.2f ms,", ave/1000/1000)
	med, _ := stats.Median(statsData)
	fmt.Fprintf(w, " Median = %.2f ms\n", med/1000/1000)

	percentile90, _ := stats.Percentile(statsData, float64(90))
	fmt.Fprintf(w, " 90%s of Request <= %.2f ms\n",
		"%%", percentile90/1000/1000)
	percentile75, _ := stats.Percentile(statsData, float64(75))
	fmt.Fprintf(w, " 75%s of Request <= %.2f ms\n",
		"%%", percentile75/1000/1000)
	percentile50, _ := stats.Percentile(statsData, float64(50))
	fmt.Fprintf(w, " 50%s of Request <= %.2f ms\n",
		"%%", percentile50/1000/1000)
	return w.String()
}
