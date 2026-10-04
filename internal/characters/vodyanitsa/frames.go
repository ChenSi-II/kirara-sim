package vodyanitsa

// User image 3, rounded to the nearest frame at 60 FPS. These are delays
// from the initial E cast, not intervals or extra animation lockouts.
// See PLACEHOLDER_FRAMES.md for provenance and remaining placeholders.
var microphoneAttackFrames = []int{228, 401, 580, 752, 934, 1106, 1285, 1457}
var microphoneHealFrames = []int{138, 222, 312, 398, 489, 574, 666, 748, 839, 924, 1017, 1101, 1190, 1275, 1364, 1451}

const (
	skillHitmark = 38
	skillLength  = 63
	burstHitmark = 102
	burstLength  = 105
)
