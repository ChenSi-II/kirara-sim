package sandrone

// Source: user images 1/6, round(seconds * 60); see PLACEHOLDER_FRAMES.md.
// CA-only C0: sweeps .816,1.166,1.516,1.816,2.133,2.466,2.800,
// 3.166,3.483,3.816; rays 1.850,2.800,3.800. The recording then
// overheats, with shots at 4.133,4.550,...,8.916 (~24f apart).
// User correction: C0 enters overdrive ON ray #3 (228f); C1 ON ray #6
// (402f). The later sweep hits at 229/409 are in-flight damage, not proof
// that resolution mode continues. A fitted gauge uses +10/s and +70/3 per
// ray, with both halved at C1, to satisfy those boundaries from zero power.
// Outside these observations, power decay/repair still need gauge footage.
// Unrecorded sweep/ray continuation is extrapolated at 20f/60f.
var resolutionSweepHitmarks = []int{49, 70, 91, 109, 128, 148, 168, 190, 209, 229}

var resolutionRayHitmarks = []int{111, 168, 228}

var resolutionC1SweepHitmarks = []int{49, 71, 93, 109, 129, 149, 169, 189, 211, 229, 250, 271, 289, 308, 329, 349, 370, 389, 409}

var resolutionC1RayHitmarks = []int{109, 169, 225, 285, 342, 402}

// E -> CA column is timed from E, not CA. E ends at .616s (37f),
// so subtract 37 from the recorded E-relative sweep/ray timestamps.
// The ES -> CA column is not used: its sprint transition is not specified.
var skillResolutionSweepHitmarks = []int{32, 51, 71, 92, 113, 134, 154, 175, 196, 214}

var skillResolutionRayHitmarks = []int{96, 154, 211}

// Subtract the confirmed 228f mode boundary from the observed C0 Z4 hits
// 248,273,296,319,343,368,392,415,439,463,486,510,535.
// Reusing these offsets after a different power history is an estimate.
var overdriveHitmarks = []int{20, 45, 68, 91, 115, 140, 164, 187, 211, 235, 258, 282, 307}

var burstBombardmentHitmarks = []int{130, 139, 151}
