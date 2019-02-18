package govppmux

import "git.fd.io/govpp.git/adapter"

// ListStats returns all stats names
func (p *Plugin) ListStats(prefixes ...string) ([]string, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.ListStats(prefixes...)
}

// DumpStats returns all stats with name, type and value
func (p *Plugin) DumpStats(prefixes ...string) ([]*adapter.StatEntry, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.DumpStats(prefixes...)
}
