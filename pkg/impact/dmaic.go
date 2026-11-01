package impact

// DMAIC readiness thresholds, shared by lint and the DMAIC export so the two
// agree on what "enough data" means.
const (
	// DMAICMinPoints is the fewest measurements that let a DMAIC effort set
	// control limits: statistical process control asks for 20 to 25 points
	// before the limits mean anything. Lint asks for this many a year; the
	// export computes control limits only from this many observations.
	DMAICMinPoints = 20
	// DMAICMinResponsesPerPoint is the fewest survey responses behind one
	// measurement before its value stops swinging with chance. An NPS needs
	// about 100 for a margin near ±10 points; 30 is the floor.
	DMAICMinResponsesPerPoint = 30
)
