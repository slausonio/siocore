package siocore

import "golang.org/x/exp/constraints"

// MapKeys is an interface that represents the possible key types for a map.
// It can be either a string or an int.
type MapKeys interface {
	constraints.Ordered
}

// MergeMaps takes in multiple maps and merges them into a single map.
// The function accepts maps with keys of type MapKeys and values of any type.
// It returns the merged map.
//
// Example Usage:
//
//	map1 := map[string]int{"a": 1, "b": 2}
//	map2 := map[string]int{"c": 3, "d": 4}
//	mergedMap := MergeMaps(map1, map2)
func MergeMaps[K MapKeys, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)

	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}

	return result
}

// MergeEnvs takes in multiple envs and merges them into a single env.
func MergeEnvs(envs ...Env) Env {
	result := make(Env)

	for _, e := range envs {
		for key, value := range e {
			result[key] = value
		}
	}

	return result
}
