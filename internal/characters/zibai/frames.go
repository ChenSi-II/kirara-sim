package zibai

// Estimated from the user's e+aqaE+3*(4aE) = approximately 12.5s rotation,
// not extracted from an image. See PLACEHOLDER_FRAMES.md (2026-10-04).
// With normal counters reset by Q/E, the action-only budget is
// 5*48 + 2*28 + 3*(28+30+36+32) + 76 = 750f, plus one 12f swap = 762f.
// Resource availability is deliberately not bypassed to enforce this budget.
// Hitmarks remain unmeasured; the existing 76f Q is also still provisional.
const phaseSkillFrames = 48

var phaseNormalFrames = [...]int{28, 30, 36, 32}
