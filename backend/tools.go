// backend/tools.go
package main

import "strings"

// ======================================================================
// Tool Registry
// ======================================================================

type ToolRegistry struct {
	api               *PokeAPIClient
	LastDetectedTypes []string
}

func newToolRegistry() *ToolRegistry {
	return &ToolRegistry{api: newPokeAPIClient()}
}

// ── declarations ──────────────────────────────────────────────────────

type ToolDeclaration struct {
	Name       string
	Desc       string
	Properties map[string]ToolProp
	Required   []string
}

type ToolProp struct {
	Type string
	Desc string
}

func (t *ToolRegistry) Declarations() []ToolDeclaration {
	return []ToolDeclaration{
		{"get_pokemon_basic_info", "Get basic info: types, stats, height, weight, generation, dex entry.",
			map[string]ToolProp{"pokemon_name": {"string", "Name or ID"}}, []string{"pokemon_name"}},
		{"get_pokemon_locations", "Get locations where a Pokémon can be found, with game version info.",
			map[string]ToolProp{"pokemon_name": {"string", "Name or ID"}}, []string{"pokemon_name"}},
		{"get_pokemon_moves", "Get moves a Pokémon learns. Optionally filter by version group.",
			map[string]ToolProp{"pokemon_name": {"string", "Name or ID"}, "game_version": {"string", "Optional version group filter"}}, []string{"pokemon_name"}},
		{"get_pokemon_evolution_chain", "Full evolution chain with conditions.",
			map[string]ToolProp{"pokemon_name": {"string", "Name or ID"}}, []string{"pokemon_name"}},
		{"get_pokemon_abilities", "All abilities including hidden, with effect descriptions.",
			map[string]ToolProp{"pokemon_name": {"string", "Name or ID"}}, []string{"pokemon_name"}},
		{"get_pokemon_by_type", "List Pokémon of a type plus type effectiveness chart.",
			map[string]ToolProp{"pokemon_type": {"string", "Type like fire water dragon"}}, []string{"pokemon_type"}},
		{"get_item_info", "Item details: category, cost, effect, flavor text.",
			map[string]ToolProp{"item_name": {"string", "e.g. thunder-stone, life-orb"}}, []string{"item_name"}},
		{"get_move_details", "Move details: type, power, accuracy, PP, effect, who learns it.",
			map[string]ToolProp{"move_name": {"string", "e.g. thunderbolt, surf"}}, []string{"move_name"}},
		{"get_nature_info", "Nature stat boosts/drops. Pass 'all' to list every nature.",
			map[string]ToolProp{"nature_name": {"string", "e.g. adamant, jolly, or all"}}, []string{"nature_name"}},
		{"search_pokemon", "Search Pokémon by name prefix.",
			map[string]ToolProp{"query": {"string", "Name prefix"}}, []string{"query"}},
	}
}

// ── Gemini payload ────────────────────────────────────────────────────

func (t *ToolRegistry) GeminiPayload() []map[string]any {
	var decls []map[string]any
	for _, d := range t.Declarations() {
		props := map[string]any{}
		for k, p := range d.Properties {
			props[k] = map[string]any{"type": p.Type, "description": p.Desc}
		}
		decls = append(decls, map[string]any{
			"name": d.Name, "description": d.Desc,
			"parameters": map[string]any{"type": "object", "properties": props, "required": d.Required},
		})
	}
	return []map[string]any{{"function_declarations": decls}}
}

// ── dispatch ──────────────────────────────────────────────────────────

func (t *ToolRegistry) Call(name string, args map[string]any) map[string]any {
	t.LastDetectedTypes = nil
	switch name {
	case "get_pokemon_basic_info":
		return t.getPokemonBasicInfo(args)
	case "get_pokemon_locations":
		return t.getPokemonLocations(args)
	case "get_pokemon_moves":
		return t.getPokemonMoves(args)
	case "get_pokemon_evolution_chain":
		return t.getPokemonEvolutionChain(args)
	case "get_pokemon_abilities":
		return t.getPokemonAbilities(args)
	case "get_pokemon_by_type":
		return t.getPokemonByType(args)
	case "get_item_info":
		return t.getItemInfo(args)
	case "get_move_details":
		return t.getMoveDetails(args)
	case "get_nature_info":
		return t.getNatureInfo(args)
	case "search_pokemon":
		return t.searchPokemon(args)
	default:
		return map[string]any{"error": "Unknown tool: " + name}
	}
}

// ── shared helpers ────────────────────────────────────────────────────

func (t *ToolRegistry) extractTypes(data map[string]any) {
	if arr, ok := data["types"].([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				if ti, ok := m["type"].(map[string]any); ok {
					if n, ok := ti["name"].(string); ok {
						t.LastDetectedTypes = append(t.LastDetectedTypes, n)
					}
				}
			}
		}
	}
}

func englishText(entries []any, field string) string {
	for _, e := range entries {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		lang, _ := m["language"].(map[string]any)
		if lang == nil || lang["name"] != "en" {
			continue
		}
		if text, ok := m[field].(string); ok {
			text = strings.ReplaceAll(text, "\n", " ")
			text = strings.ReplaceAll(text, "\f", " ")
			return text
		}
	}
	return ""
}

func mapName(m map[string]any) string {
	if n, ok := m["name"].(string); ok {
		return n
	}
	return "?"
}

func unique(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// ======================================================================
// Tool implementations (unchanged logic)
// ======================================================================

func (t *ToolRegistry) getPokemonBasicInfo(args map[string]any) map[string]any {
	name := strArg(args, "pokemon_name")
	if name == "" {
		return map[string]any{"error": "Missing pokemon_name"}
	}
	data, err := t.api.Get("pokemon/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	t.extractTypes(data)
	generation, flavor := "unknown", ""
	if spec, ok := data["species"].(map[string]any); ok {
		if u, ok := spec["url"].(string); ok {
			if sp, err := t.api.GetURL(u); err == nil {
				if gen, ok := sp["generation"].(map[string]any); ok {
					generation = mapName(gen)
				}
				if fte, ok := sp["flavor_text_entries"].([]any); ok {
					flavor = englishText(fte, "flavor_text")
				}
			}
		}
	}
	stats := map[string]any{}
	if arr, ok := data["stats"].([]any); ok {
		for _, s := range arr {
			sm := s.(map[string]any)
			si := sm["stat"].(map[string]any)
			stats[mapName(si)] = sm["base_stat"]
		}
	}
	var types []string
	if arr, ok := data["types"].([]any); ok {
		for _, ti := range arr {
			tm := ti.(map[string]any)
			if tp, ok := tm["type"].(map[string]any); ok {
				types = append(types, mapName(tp))
			}
		}
	}
	return map[string]any{"name": data["name"], "id": data["id"], "types": types,
		"height_decimeters": data["height"], "weight_hectograms": data["weight"],
		"base_stats": stats, "generation": generation,
		"base_experience": data["base_experience"], "flavor_text": flavor}
}

func (t *ToolRegistry) getPokemonLocations(args map[string]any) map[string]any {
	name := strArg(args, "pokemon_name")
	if name == "" {
		return map[string]any{"error": "Missing pokemon_name"}
	}
	data, err := t.api.Get("pokemon/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	t.extractTypes(data)
	encURL, ok := data["location_area_encounters"].(string)
	if !ok || encURL == "" {
		return map[string]any{"name": name, "locations": []any{}, "note": "No encounter data."}
	}
	arr, err := t.api.GetArray(encURL)
	if err != nil {
		return map[string]any{"name": name, "locations": []any{}, "note": err.Error()}
	}
	var locs []map[string]any
	for _, loc := range arr {
		area, ok := loc["location_area"].(map[string]any)
		if !ok {
			continue
		}
		aName := strings.ReplaceAll(mapName(area), "-", " ")
		versions := map[string][]string{}
		if vds, ok := loc["version_details"].([]any); ok {
			for _, vd := range vds {
				vdm := vd.(map[string]any)
				vi := vdm["version"].(map[string]any)
				vn := mapName(vi)
				var methods []string
				if eds, ok := vdm["encounter_details"].([]any); ok {
					for _, ed := range eds {
						edm := ed.(map[string]any)
						if m, ok := edm["method"].(map[string]any); ok {
							methods = append(methods, mapName(m))
						}
					}
				}
				versions[vn] = unique(methods)
			}
		}
		locs = append(locs, map[string]any{"area": aName, "versions": versions})
	}
	return map[string]any{"name": name, "locations": locs}
}

func (t *ToolRegistry) getPokemonMoves(args map[string]any) map[string]any {
	name := strArg(args, "pokemon_name")
	if name == "" {
		return map[string]any{"error": "Missing pokemon_name"}
	}
	vf := strArg(args, "game_version")
	if vf != "" {
		vf = norm(vf)
	}
	data, err := t.api.Get("pokemon/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	t.extractTypes(data)
	moveArr, ok := data["moves"].([]any)
	if !ok {
		return map[string]any{"name": name, "moves_by_learn_method": map[string]any{}}
	}
	byMethod := map[string][]map[string]any{}
	for _, me := range moveArr {
		mem := me.(map[string]any)
		mi := mem["move"].(map[string]any)
		mn := mapName(mi)
		vgds, ok := mem["version_group_details"].([]any)
		if !ok {
			continue
		}
		for _, vgd := range vgds {
			vgdm := vgd.(map[string]any)
			vg := vgdm["version_group"].(map[string]any)
			vgn := mapName(vg)
			if vf != "" && !strings.Contains(vgn, vf) {
				continue
			}
			method := mapName(vgdm["move_learn_method"].(map[string]any))
			byMethod[method] = append(byMethod[method], map[string]any{
				"name": mn, "level": vgdm["level_learned_at"], "version_group": vgn,
			})
		}
	}
	filter := "all"
	if vf != "" {
		filter = vf
	}
	return map[string]any{"name": name, "version_filter": filter, "moves_by_learn_method": byMethod}
}

func (t *ToolRegistry) getPokemonEvolutionChain(args map[string]any) map[string]any {
	name := strArg(args, "pokemon_name")
	if name == "" {
		return map[string]any{"error": "Missing pokemon_name"}
	}
	if pData, err := t.api.Get("pokemon/"+norm(name), nil); err == nil {
		t.extractTypes(pData)
	}
	species, err := t.api.Get("pokemon-species/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	ci, ok := species["evolution_chain"].(map[string]any)
	if !ok {
		return map[string]any{"name": name, "chain": "No evolution chain data."}
	}
	cu, ok := ci["url"].(string)
	if !ok {
		return map[string]any{"name": name, "chain": "No chain URL."}
	}
	cd, err := t.api.GetURL(cu)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	chain, ok := cd["chain"].(map[string]any)
	if !ok {
		return map[string]any{"name": name, "chain": "Parse error."}
	}
	return map[string]any{"name": name, "evolution_chain": parseChain(chain)}
}

func parseChain(node map[string]any) map[string]any {
	sn := "?"
	if sp, ok := node["species"].(map[string]any); ok {
		sn = mapName(sp)
	}
	var evos []map[string]any
	if et, ok := node["evolves_to"].([]any); ok {
		for _, e := range et {
			evo := e.(map[string]any)
			tn := "?"
			if sp, ok := evo["species"].(map[string]any); ok {
				tn = mapName(sp)
			}
			var conds []map[string]any
			if ds, ok := evo["evolution_details"].([]any); ok {
				for _, d := range ds {
					dm := d.(map[string]any)
					c := map[string]any{}
					if v, ok := dm["min_level"].(float64); ok && v > 0 {
						c["min_level"] = int(v)
					}
					if v, ok := dm["item"].(map[string]any); ok {
						c["item"] = mapName(v)
					}
					if v, ok := dm["trigger"].(map[string]any); ok {
						c["trigger"] = mapName(v)
					}
					if v, ok := dm["held_item"].(map[string]any); ok {
						c["held_item"] = mapName(v)
					}
					if v, ok := dm["min_happiness"].(float64); ok && v > 0 {
						c["min_happiness"] = int(v)
					}
					if v, ok := dm["known_move"].(map[string]any); ok {
						c["known_move"] = mapName(v)
					}
					if v, ok := dm["time_of_day"].(string); ok && v != "" {
						c["time_of_day"] = v
					}
					if len(c) > 0 {
						conds = append(conds, c)
					}
				}
			}
			evos = append(evos, map[string]any{
				"into": tn, "conditions": conds, "further": parseChain(evo),
			})
		}
	}
	return map[string]any{"species": sn, "evolves_to": evos}
}

func (t *ToolRegistry) getPokemonAbilities(args map[string]any) map[string]any {
	name := strArg(args, "pokemon_name")
	if name == "" {
		return map[string]any{"error": "Missing pokemon_name"}
	}
	data, err := t.api.Get("pokemon/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	t.extractTypes(data)
	arr, ok := data["abilities"].([]any)
	if !ok {
		return map[string]any{"name": name, "abilities": []any{}}
	}
	var abilities []map[string]any
	for _, slot := range arr {
		sm := slot.(map[string]any)
		info := sm["ability"].(map[string]any)
		an := mapName(info)
		hidden, _ := sm["is_hidden"].(bool)
		effect := ""
		if ad, err := t.api.Get("ability/"+an, nil); err == nil {
			if ee, ok := ad["effect_entries"].([]any); ok {
				effect = englishText(ee, "short_effect")
			}
		}
		abilities = append(abilities, map[string]any{"name": an, "is_hidden": hidden, "effect": effect})
	}
	return map[string]any{"name": name, "abilities": abilities}
}

func (t *ToolRegistry) getPokemonByType(args map[string]any) map[string]any {
	tn := strArg(args, "pokemon_type")
	if tn == "" {
		return map[string]any{"error": "Missing pokemon_type"}
	}
	t.LastDetectedTypes = []string{norm(tn)}
	data, err := t.api.Get("type/"+norm(tn), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var list []map[string]any
	if arr, ok := data["pokemon"].([]any); ok {
		for i, e := range arr {
			if i >= 50 {
				break
			}
			em := e.(map[string]any)
			if p, ok := em["pokemon"].(map[string]any); ok {
				list = append(list, map[string]any{"name": mapName(p)})
			}
		}
	}
	eff := map[string][]string{}
	if dr, ok := data["damage_relations"].(map[string]any); ok {
		for k, v := range dr {
			if a, ok := v.([]any); ok {
				var names []string
				for _, item := range a {
					if m, ok := item.(map[string]any); ok {
						names = append(names, mapName(m))
					}
				}
				eff[k] = names
			}
		}
	}
	return map[string]any{"type": tn, "pokemon_sample": list, "type_effectiveness": eff}
}

func (t *ToolRegistry) getItemInfo(args map[string]any) map[string]any {
	name := strArg(args, "item_name")
	if name == "" {
		return map[string]any{"error": "Missing item_name"}
	}
	data, err := t.api.Get("item/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	effect, flavor := "", ""
	if ee, ok := data["effect_entries"].([]any); ok {
		effect = englishText(ee, "effect")
	}
	if fe, ok := data["flavor_text_entries"].([]any); ok {
		flavor = englishText(fe, "flavor_text")
	}
	cat := "unknown"
	if c, ok := data["category"].(map[string]any); ok {
		cat = mapName(c)
	}
	return map[string]any{"name": data["name"], "category": cat, "cost": data["cost"],
		"effect": effect, "flavor_text": flavor}
}

func (t *ToolRegistry) getMoveDetails(args map[string]any) map[string]any {
	name := strArg(args, "move_name")
	if name == "" {
		return map[string]any{"error": "Missing move_name"}
	}
	data, err := t.api.Get("move/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	moveType := "unknown"
	if tp, ok := data["type"].(map[string]any); ok {
		moveType = mapName(tp)
	}
	t.LastDetectedTypes = []string{moveType}
	effect := ""
	if ee, ok := data["effect_entries"].([]any); ok {
		effect = englishText(ee, "effect")
	}
	var learnedBy []string
	if arr, ok := data["learned_by_pokemon"].([]any); ok {
		for i, p := range arr {
			if i >= 20 {
				break
			}
			if pm, ok := p.(map[string]any); ok {
				learnedBy = append(learnedBy, mapName(pm))
			}
		}
	}
	dc := "unknown"
	if d, ok := data["damage_class"].(map[string]any); ok {
		dc = mapName(d)
	}
	return map[string]any{"name": data["name"], "type": moveType, "power": data["power"],
		"accuracy": data["accuracy"], "pp": data["pp"], "damage_class": dc,
		"effect": effect, "learned_by_sample": learnedBy}
}

func (t *ToolRegistry) getNatureInfo(args map[string]any) map[string]any {
	name := strArg(args, "nature_name")
	if name == "" {
		return map[string]any{"error": "Missing nature_name"}
	}
	if norm(name) == "all" {
		data, err := t.api.Get("nature", map[string]string{"limit": "25"})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		results, ok := data["results"].([]any)
		if !ok {
			return map[string]any{"error": "Failed to parse natures"}
		}
		var natures []map[string]any
		for _, r := range results {
			rm := r.(map[string]any)
			nn := mapName(rm)
			if nd, err := t.api.Get("nature/"+nn, nil); err == nil {
				inc, dec := "none", "none"
				if is, ok := nd["increased_stat"].(map[string]any); ok {
					inc = mapName(is)
				}
				if ds, ok := nd["decreased_stat"].(map[string]any); ok {
					dec = mapName(ds)
				}
				natures = append(natures, map[string]any{"name": nn, "increases": inc, "decreases": dec})
			}
		}
		return map[string]any{"natures": natures}
	}
	data, err := t.api.Get("nature/"+norm(name), nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	inc, dec, likes, hates := "none", "none", "none", "none"
	if is, ok := data["increased_stat"].(map[string]any); ok {
		inc = mapName(is)
	}
	if ds, ok := data["decreased_stat"].(map[string]any); ok {
		dec = mapName(ds)
	}
	if lf, ok := data["likes_flavor"].(map[string]any); ok {
		likes = mapName(lf)
	}
	if hf, ok := data["hates_flavor"].(map[string]any); ok {
		hates = mapName(hf)
	}
	return map[string]any{"name": data["name"], "increases": inc, "decreases": dec,
		"likes_flavor": likes, "dislikes_flavor": hates}
}

func (t *ToolRegistry) searchPokemon(args map[string]any) map[string]any {
	q := strArg(args, "query")
	if q == "" {
		return map[string]any{"error": "Missing query"}
	}
	prefix := norm(q)
	data, err := t.api.Get("pokemon", map[string]string{"limit": "1302"})
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	results, ok := data["results"].([]any)
	if !ok {
		return map[string]any{"error": "Failed to list pokémon"}
	}
	var matches []map[string]any
	for _, r := range results {
		rm := r.(map[string]any)
		n := mapName(rm)
		if strings.HasPrefix(n, prefix) {
			matches = append(matches, map[string]any{"name": n})
		}
	}
	total := len(matches)
	if len(matches) > 20 {
		matches = matches[:20]
	}
	return map[string]any{"query": q, "matches": matches, "total": total}
}
