package vesna

// Timing provenance: user-provided image 2, converted at 60 FPS; see
// PLACEHOLDER_FRAMES.md. Only the armed-state normal string was measured.
// Its absolute hitmarks are 20, 45, 77, 93, 123, 153, 208, ending at 227.
// Per-action boundaries are inferred by assigning a 20-frame first hit to
// each normal, not independently measured cancel points.
var armedNormalHitmarks = [][]int{{20}, {20}, {20, 36}, {20}, {20}, {20}}
var armedNormalLengths = []int{25, 32, 46, 30, 55, 39}

const (
	// A1 -> charged attack is inferred as a 20-frame A1 plus 58-frame CA:
	// the image gives A1 at 20, CA at 58, and the combined action ending at 78.
	armedNormalChargeCancel = 20
	armedChargeHitmark      = 38
	armedChargeLength       = 58
	// Feather timings in the same A1+CA sample: 73, 115, 115 absolute.
	armedNormalFeatherDelay = 53 // after the normal hit; extrapolated to N2-N6
	armedChargeFeather      = 95 // relative to the inferred CA start at frame 20
	initialSkillHitmark     = 33
	initialSkillLength      = 27
	burstHitmark            = 135
	burstLength             = 133
	burstToDance            = 131
)

var spiritbladeSkillLengths = []int{20, 52, 76}
