package config

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// isMaxBinding — возвращает true, если первый аргумент больше всех остальных.
func isMaxBinding(args ...ref.Val) ref.Val {
	if len(args) < 2 {
		return types.Bool(false)
	}
	first, ok := args[0].Value().(int64)
	if !ok {
		return types.Bool(false)
	}
	for _, v := range args[1:] {
		val, ok := v.Value().(int64)
		if !ok || first <= val {
			return types.Bool(false)
		}
	}
	return types.Bool(true)
}

// IsMaxLib — EnvOption, регистрирующий функцию isMax для CEL.
var IsMaxLib = cel.Function(
	"isMax",
	cel.Overload(
		"isMax_int_varargs_bool",
		[]*cel.Type{cel.IntType, cel.DynType},
		cel.BoolType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			return isMaxBinding(args...)
		}),
	),
)
