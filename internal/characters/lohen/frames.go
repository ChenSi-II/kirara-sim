package lohen

// Source: user timing image 8, converted at 60 FPS (PLACEHOLDER_FRAMES.md).
// The E -> N5 recording has hits at 33,55,81,93,105,119,155,177
// and ends at 207, with the first normal starting at 21. Individual normal
// boundaries are inferred using a 12-frame initial hit; they are not measured
// per-action cancel windows. Only the Masterstroke stance uses this table.
var masterstrokeNormalHitmarks = [][]int{{12}, {12}, {12, 24, 36}, {12}, {12, 34}}

var masterstrokeNormalFrames = []int{22, 26, 38, 36, 64}

// N1 -> CA uses the N1 hit at 12 as an inferred transition. The resulting
// N1C hits are at 12,37,45 and repeat in about 48 frames, as in image 8.
var masterstrokeChargeHitmarks = []int{25, 33}

var etchedHitmarks = []int{27, 34, 45, 50}

var burstHitmarks = []int{106, 131, 137, 140, 147, 171}
