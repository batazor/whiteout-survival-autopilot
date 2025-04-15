package heroes

// BestForResource возвращает список доступных героев с баффом на указанный ресурс.
func (h Heroes) BestForResource(resource string) []Hero {
	var result []Hero
	key := resource + "_gathering_speed"

	for _, hero := range h {
		if !hero.State.IsAvailable {
			continue
		}
		if _, ok := hero.Buffs[key]; ok {
			result = append(result, hero)
		}
	}
	return result
}

// BestForDefense возвращает доступных героев с ролью обороны.
func (h Heroes) BestForDefense() []Hero {
	var result []Hero
	for _, hero := range h {
		if !hero.State.IsAvailable {
			continue
		}
		for _, role := range hero.Roles {
			if role == "garrison_defense" || role == "defense" {
				result = append(result, hero)
				break
			}
		}
	}
	return result
}

// BestForAttack возвращает доступных героев с боевой ролью.
func (h Heroes) BestForAttack() []Hero {
	var result []Hero
	for _, hero := range h {
		if !hero.State.IsAvailable {
			continue
		}
		for _, role := range hero.Roles {
			if role == "rally_leader" || role == "combat" {
				result = append(result, hero)
				break
			}
		}
	}
	return result
}

func (h Heroes) Available() Heroes {
	out := Heroes{}
	for name, hero := range h {
		if hero.State.IsAvailable {
			out[name] = hero
		}
	}
	return out
}
