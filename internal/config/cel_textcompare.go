package config

import (
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

// FuzzySubstringMatch возвращает true, если needle встречается в haystack
// c максимум maxDist ошибками Левенштейна внутри любого окна той же длины.
// Дополнительно: если haystack короче needle, сравниваем строки целиком.
func FuzzySubstringMatch(haystack, needle string, maxDist int) bool {
	haystack = strings.ToLower(haystack)
	needle = strings.ToLower(needle)

	n, m := len(needle), len(haystack)
	if n == 0 {
		return false
	}

	// haystack короче needle → сравниваем целиком
	if m < n {
		return levenshtein.ComputeDistance(haystack, needle) <= maxDist
	}

	for i := 0; i <= m-n; i++ {
		if levenshtein.ComputeDistance(haystack[i:i+n], needle) <= maxDist {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// compareText(a, b) → bool  (регистрация в CEL)
// -----------------------------------------------------------------------------

// compareTextBinding — фактическая реализация функции для CEL.
func compareTextBinding(lhs, rhs ref.Val) ref.Val {
	a, ok1 := lhs.Value().(string)
	b, ok2 := rhs.Value().(string)
	if !ok1 || !ok2 {
		return types.Bool(false)
	}

	al := strings.ToLower(a)
	bl := strings.ToLower(b)

	if strings.Contains(bl, al) || FuzzySubstringMatch(bl, al, 1) {
		return types.Bool(true)
	}
	return types.Bool(false)
}

// CompareTextLib — EnvOption, регистрирующий функцию compareText.
var CompareTextLib = cel.Function(
	"compareText",
	cel.Overload(
		"compareText_string_string_bool",
		[]*cel.Type{cel.StringType, cel.StringType},
		cel.BoolType,
		cel.BinaryBinding(compareTextBinding),
	),
)
