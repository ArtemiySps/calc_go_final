package models

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// функция для определения приоритета оператора
func precedence(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	default:
		return 0
	}
}

// функция для преобразования выражения в постфиксную запись (обратная польская запись)
func InfixToPostfix(expression string) (string, error) {
	var stack []rune
	var output strings.Builder
	var numBuffer strings.Builder

	for _, char := range expression {
		switch {
		case unicode.IsDigit(char):
			numBuffer.WriteRune(char)

		case char == '(':
			if numBuffer.Len() > 0 {
				output.WriteString(numBuffer.String())
				output.WriteRune(' ')
				numBuffer.Reset()
			}
			stack = append(stack, char)

		case char == ')':
			if numBuffer.Len() > 0 {
				output.WriteString(numBuffer.String())
				output.WriteRune(' ')
				numBuffer.Reset()
			}

			for len(stack) > 0 && stack[len(stack)-1] != '(' {
				output.WriteRune(stack[len(stack)-1])
				output.WriteRune(' ')
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return "", ErrBadExpression
			}
			stack = stack[:len(stack)-1]

		case char == '+' || char == '-' || char == '*' || char == '/':
			if numBuffer.Len() > 0 {
				output.WriteString(numBuffer.String())
				output.WriteRune(' ')
				numBuffer.Reset()
			}

			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(char) {
				output.WriteRune(stack[len(stack)-1])
				output.WriteRune(' ')
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, char)

		default:
			return "", ErrUnexpectedSymbol
		}
	}

	if numBuffer.Len() > 0 {
		output.WriteString(numBuffer.String())
		output.WriteRune(' ')
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == '(' {
			return "", ErrBadExpression
		}
		output.WriteRune(stack[len(stack)-1])
		output.WriteRune(' ')
		stack = stack[:len(stack)-1]
	}

	return strings.TrimSpace(output.String()), nil
}

// функция для создания ID
func MakeID() string {
	return uuid.New().String()
}

// функция создания zap.Logger
func MakeLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
