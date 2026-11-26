package checker

func (cfg Config) Check(edges []Edge, onFile func(string)) []violation {
	layers := make(map[string]int)
	layer := func(file string) int {
		if index, known := layers[file]; known {
			return index
		}
		index := cfg.layer(file)
		layers[file] = index
		if cfg.debug {
			onFile(file)
		}
		return index
	}
	violations := []violation{}
	for _, e := range edges {
		from, to := layer(e.From), layer(e.To)
		if from >= 0 && from < to {
			violations = append(violations, violation{e, cfg.layers[from], cfg.layers[to]})
		}
	}
	return violations
}
