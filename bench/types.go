package bench

type Run struct {
	Version string   `json:"version,omitempty"`
	Date    int64    `json:"date,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Suites  []Suite  `json:"suites"`
}

type Suite struct {
	Go         string      `json:"go,omitempty"`
	Goos       string      `json:"goos"`
	Goarch     string      `json:"goarch"`
	ShortPath  string      `json:"short_path,omitempty"`
	Pkg        string      `json:"pkg"`
	Benchmarks []Benchmark `json:"benchmarks"`
}

type Benchmark struct {
	Name    string             `json:"name"`
	Runs    int64              `json:"runs"`
	NsPerOp float64            `json:"nsPerOp,omitempty"`
	Mem     *Mem               `json:"mem,omitempty"`
	Custom  map[string]float64 `json:"custom,omitempty"`
}

type Mem struct {
	BytesPerOp  float64 `json:"bytesPerOp,omitempty"`
	AllocsPerOp float64 `json:"allocsPerOp,omitempty"`
	MBPerSec    float64 `json:"mbPerSec,omitempty"`
}

type ComparisonResult struct {
	Name        string
	OldRuns     int64
	NewRuns     int64
	OldNsPerOp  float64
	NewNsPerOp  float64
	NsPerOpDiff float64
	NsPerOpPct  float64
	OldBytes    float64
	NewBytes    float64
	BytesDiff   float64
	BytesPct    float64
	OldAllocs   float64
	NewAllocs   float64
	AllocsDiff  float64
	AllocsPct   float64
}

func (c *ComparisonResult) IsRegression(threshold float64) bool {

	if c.NsPerOpPct > threshold {
		return true
	}

	if c.BytesPct > threshold {
		return true
	}

	if c.AllocsPct > threshold {
		return true
	}

	return false
}
