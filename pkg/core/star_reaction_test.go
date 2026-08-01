package core

import "testing"

func TestIsStarReactionAvatar(t *testing.T) {
	tests := []struct {
		name  string
		id    int32
		subID int32
		want  bool
	}{
		{name: "avatar 10000150", id: 10000150, want: true},
		{name: "avatar 10000133", id: 10000133, want: true},
		{name: "cryo lumine placeholder", id: 10000007, subID: 705, want: true},
		{name: "anemo lumine is not eligible", id: 10000007, subID: 704, want: false},
		{name: "unlisted avatar", id: 10000131, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isStarReactionAvatar(tc.id, tc.subID); got != tc.want {
				t.Fatalf("isStarReactionAvatar(%d, %d) = %v, want %v", tc.id, tc.subID, got, tc.want)
			}
		})
	}
}
