package dot

import "fmt"

// packWithinBudget reserves the zero-selection manifest, then admits candidates
// in caller-defined priority order only when the fully rendered payload still fits.
func packWithinBudget(maxSize, candidates int, render func([]bool) []byte) ([]byte, error) {
	selected := make([]bool, candidates)
	payload := render(selected)
	if len(payload) > maxSize {
		return nil, fmt.Errorf("budget %d bytes cannot hold the %d-byte omission manifest", maxSize, len(payload))
	}
	for candidate := range candidates {
		selected[candidate] = true
		trial := render(selected)
		if len(trial) <= maxSize {
			payload = trial
		} else {
			selected[candidate] = false
		}
	}
	return payload, nil
}
