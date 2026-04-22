package generator

import "fmt"

const (
	StackViteReactTailwind = "vite-react-tailwind"
	StackNextTailwind      = "next-tailwind"
	StackStaticHTMLCSS     = "static-html-css"
)

func ValidateStack(stack string) error {
	switch stack {
	case StackViteReactTailwind, StackNextTailwind, StackStaticHTMLCSS:
		return nil
	default:
		return fmt.Errorf("unsupported stack %q", stack)
	}
}

func IsImplementedStack(stack string) bool {
	return stack == StackViteReactTailwind
}
