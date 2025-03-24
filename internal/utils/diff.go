package utils

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/lipgloss"
)

// PrintStyledDiff takes two pointers to structs and prints their diff using colors
func PrintStyledDiff(oldVal, newVal interface{}) {
	oldV := reflect.ValueOf(oldVal).Elem()
	newV := reflect.ValueOf(newVal).Elem()
	typ := oldV.Type()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	oldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	newStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	fmt.Println(titleStyle.Render("🔍 State Changes:"))

	compareStructFields(oldV, newV, typ, "", fieldStyle, oldStyle, newStyle)
}

func compareStructFields(oldV, newV reflect.Value, typ reflect.Type, prefix string, fieldStyle, oldStyle, newStyle lipgloss.Style) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := field.Name
		oldField := oldV.Field(i)
		newField := newV.Field(i)

		fieldPath := name
		if prefix != "" {
			fieldPath = prefix + "." + name
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			compareStructFields(oldField, newField, field.Type, fieldPath, fieldStyle, oldStyle, newStyle)
			continue
		}

		// Compare values
		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			fmt.Printf(" %s:\n", fieldStyle.Render(fieldPath))
			fmt.Printf("   old → %s\n", oldStyle.Render(fmt.Sprintf("%v", oldField.Interface())))
			fmt.Printf("   new → %s\n", newStyle.Render(fmt.Sprintf("%v", newField.Interface())))
		}
	}
}
