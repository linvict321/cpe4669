package main

//maps node IDs to counters
type VectorClock map[int]int

func(vc VectorClock) Increment(nodeID int) {
	vc[nodeID]++
}

//return 1 if vc is newer, -1 if older, and 0 if equal/concurrent
func(vc VectorClock) Compare(other VectorClock) int {
	vcNewer := false
	vcOlder := false

	for i, j := range vc {
		otherCount := other[i]
		if j > otherCount {
			vcNewer = true
		} else if j < otherCount {
			vcOlder = true
		}
	}
	
	for i, j := range other {
		if selfCount, ok := vc[i]; !ok || selfCount < j {
			if j > 0 {
				vcOlder = true
			}
		}
	}

	switch {
	case vcNewer && !vcOlder:
		return 1
	case vcOlder && !vcNewer:
		return -1
	default:
		return 0
	}
}

//checks if two clocks are identical
func(vc VectorClock) Equal(other VectorClock) bool {
	if len(vc) != len(other) {
		return false
	}
	for i, j := range vc {
		if other[i] != j {
			return false
		}
	}
	return true
}

//combines two clocks by taking the highest counter for each node
func(vc VectorClock) Merge(other VectorClock) VectorClock {
	out := make(VectorClock, len(vc)+len(other))
	for i, j := range vc {
		out[i] = j
	}
	for i, j := range other {
		if j > out[i] {
			out[i] = j
		}
	}
	return out
}

type Version struct {
	Value string
	Clock VectorClock
}

type Store struct {
	data map[string][]Version
}

func NewStore() *Store {
	return &Store{data: make(map[string][]Version)}
}

//stores a value if it's new, rejects it if stale, keeps it if concurrent
func(s *Store) PutVersion(key string, value string, clock VectorClock) bool {
	siblings := s.data[key]
	if len(siblings) == 0 {
		s.data[key] = []Version{{Value: value, Clock: clock}}
		return true
	}

	stale := false
	kept := make([]Version, 0, len(siblings)+1)
	for _, sib := range siblings {
		switch clock.Compare(sib.Clock) {
		case 1:
			continue //new is newer, drop the old
		case -1:
			stale = true //new is older, reject
		case 0:
			kept = append(kept, sib) //concurrent, keep both
		}
	}

	if stale {
		return false
	}
	kept = append(kept, Version{Value: value, Clock: clock})
	s.data[key] = kept
	return true
}

func(s *Store) GetVersions(key string) []Version {
	out := make([]Version, len(s.data[key]))
	copy(out, s.data[key])
	return out
}