package core

import "strings"

// flat3: Two-phase approach (cartesian product, then process)
func flat3(frames [][]string, groups [][]string) [][]string {
	if len(groups) == 0 {
		return frames
	}

	expressions := make([][]string, 0, len(frames)*4)
	groupOffset := 0

	for _, frame := range frames {
		// Count placeholders
		subsTotal := 0
		for _, tok := range frame {
			if tok == "§" {
				subsTotal++
			}
		}

		if subsTotal == 0 {
			expressions = append(expressions, frame)
			continue
		}

		// Phase 1: Generate pure cartesian product
		divisors := make([]int, subsTotal)
		divisors[0] = 1
		for i := 1; i < subsTotal; i++ {
			divisors[i] = divisors[i-1] * len(groups[groupOffset+i-1])
		}
		totalCombos := divisors[subsTotal-1] * len(groups[groupOffset+subsTotal-1])

		combinations := make([][]string, totalCombos)
		for i := 0; i < totalCombos; i++ {
			combo := make([]string, subsTotal)
			for j := 0; j < subsTotal; j++ {
				idx := (i / divisors[j]) % len(groups[groupOffset+j])
				combo[j] = groups[groupOffset+j][idx]
			}
			combinations[i] = combo
		}

		// Phase 2: Process each combination
		for _, combo := range combinations {
			result := buildExpression(frame, combo)
			expressions = append(expressions, result...)
		}

		groupOffset += subsTotal
	}

	return expressions
}

// Helper for flatA
func buildExpression(frame []string, combo []string) [][]string {
	results := [][]string{}
	template := []string{}
	comboIdx := 0

	for _, tok := range frame {
		if tok != "§" {
			template = append(template, tok)
			continue
		}

		val := combo[comboIdx]
		comboIdx++

		// Handle special cases
		if val == "&" {
			if len(template) > 0 {
				results = append(results, template)
				template = []string{}
			}
			results = append(results, []string{"&"})
			continue
		}

		template = append(template, val)
	}

	if len(template) > 0 {
		results = append(results, template)
	}

	return results
}

// flat4: Pre-normalize groups approach
func flat4(frames [][]string, groups [][]string) [][]string {
	if len(groups) == 0 {
		return frames
	}

	expressions := make([][]string, 0, len(frames)*4)
	groupOffset := 0

	for _, frame := range frames {
		// Count placeholders
		subsTotal := 0
		subsPositions := make([]int, 0, 4)
		for i, tok := range frame {
			if tok == "§" {
				subsPositions = append(subsPositions, i)
				subsTotal++
			}
		}

		if subsTotal == 0 {
			expressions = append(expressions, frame)
			continue
		}

		// Pre-normalize groups (collapse joins, etc)
		normalizedGroups := make([][]string, subsTotal)
		for i := 0; i < subsTotal; i++ {
			group := groups[groupOffset+i]
			normalized := make([]string, 0, len(group))

			for j := 0; j < len(group); j++ {
				val := group[j]
				// Collapse joined sequences
				if val == "@" || val == "$" {
					for j < len(group)-1 {
						next := group[j+1]
						val += " " + next
						j++
						if next == ")" || next == "}" {
							break
						}
					}
				}
				normalized = append(normalized, val)
			}
			normalizedGroups[i] = normalized
		}

		// Simple iteration (no special cases!)
		divisors := make([]int, subsTotal)
		divisors[0] = 1
		for i := 1; i < subsTotal; i++ {
			divisors[i] = divisors[i-1] * len(normalizedGroups[i-1])
		}
		totalCombos := divisors[subsTotal-1] * len(normalizedGroups[subsTotal-1])

		for i := 0; i < totalCombos; i++ {
			template := append([]string(nil), frame...)

			for j := 0; j < subsTotal; j++ {
				idx := (i / divisors[j]) % len(normalizedGroups[j])
				val := normalizedGroups[j][idx]

				// Handle & separator
				if val == "&" {
					// Skip this combo, output & separately
					expressions = append(expressions, []string{"&"})
					goto nextCombo
				}

				// Split joined sequences back to tokens
				if strings.Contains(val, " ") {
					tokens := strings.Fields(val)
					// Replace placeholder with tokens
					before := template[:subsPositions[j]]
					after := template[subsPositions[j]+1:]
					template = append(before, append(tokens, after...)...)
					// Adjust positions for subsequent placeholders
					for k := j + 1; k < subsTotal; k++ {
						subsPositions[k] += len(tokens) - 1
					}
				} else {
					template[subsPositions[j]] = val
				}
			}

			expressions = append(expressions, template)
		nextCombo:
		}

		groupOffset += subsTotal
	}

	return expressions
}

// flat5: Recursive approach
func flat5(frames [][]string, groups [][]string) [][]string {
	if len(groups) == 0 {
		return frames
	}

	expressions := make([][]string, 0, len(frames)*4)
	groupOffset := 0

	for _, frame := range frames {
		// Count placeholders
		subsTotal := 0
		for _, tok := range frame {
			if tok == "§" {
				subsTotal++
			}
		}

		if subsTotal == 0 {
			expressions = append(expressions, frame)
			continue
		}

		// Recursive combination generator
		var recurse func(pos int, current []string)
		recurse = func(pos int, current []string) {
			if pos == subsTotal {
				// Build expression from current combination
				template := []string{}
				currentIdx := 0

				for _, tok := range frame {
					if tok != "§" {
						template = append(template, tok)
						continue
					}

					val := current[currentIdx]
					currentIdx++

					if val == "&" {
						if len(template) > 0 {
							expressions = append(expressions, template)
							template = []string{}
						}
						expressions = append(expressions, []string{"&"})
						return
					}

					template = append(template, val)
				}

				if len(template) > 0 {
					expressions = append(expressions, template)
				}
				return
			}

			// Recurse through group items
			group := groups[groupOffset+pos]
			for _, val := range group {
				// Collapse joined sequences
				if val == "@" || val == "$" {
					joined := val
					i := 0
					for _, item := range group {
						if i == 0 {
							i++
							continue
						}
						joined += " " + item
						if item == ")" || item == "}" {
							break
						}
						i++
					}
					recurse(pos+1, append(current, joined))
					return
				}

				recurse(pos+1, append(current, val))
			}
		}

		recurse(0, []string{})
		groupOffset += subsTotal
	}

	return expressions
}
