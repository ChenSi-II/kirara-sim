package keys

import (
	"encoding/json"
	"testing"
)

func TestWeaponNamesMatchConstants(t *testing.T) {
	if got, want := len(weaponNames), int(GoldenFrostboundOath)+1; got != want {
		t.Fatalf("weapon name/constant count mismatch: names=%d constants=%d", got, want)
	}
	for key := NoWeapon; key <= GoldenFrostboundOath; key++ {
		data, err := json.Marshal(&key)
		if err != nil {
			t.Fatalf("marshal %d: %v", key, err)
		}
		var roundTrip Weapon
		if err := json.Unmarshal(data, &roundTrip); err != nil {
			t.Fatalf("unmarshal %d (%s): %v", key, data, err)
		}
		if roundTrip != key {
			t.Fatalf("weapon key changed during JSON round trip: got=%d want=%d", roundTrip, key)
		}
	}
}
