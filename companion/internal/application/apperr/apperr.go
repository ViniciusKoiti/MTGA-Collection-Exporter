// Package apperr define os códigos estáveis de erro da aplicação do
// companion e o mapeamento seguro para apresentação (tarefa 2.4): a UI e
// o assistente veem código + operação, nunca caminhos crus ou exceções.
package apperr

import (
	"errors"
	"fmt"
)

// Code é um código estável de erro; automação e UI dependem dele, nunca
// do texto da mensagem.
type Code string

// Códigos estáveis da aplicação.
const (
	CodeSourceUnavailable  Code = "source_unavailable"
	CodeCatalogUnavailable Code = "catalog_unavailable"
	CodeSnapshotNotFound   Code = "snapshot_not_found"
	CodeValidationFailed   Code = "validation_failed"
	CodeApprovalRequired   Code = "approval_required"
	CodeApprovalDenied     Code = "approval_denied"
	CodeConflict           Code = "conflict"
	CodeAssistantDisabled  Code = "assistant_disabled"
	CodeInternal           Code = "internal"
)

// Error carrega código, operação e a causa técnica encapsulada. A causa
// aparece apenas em logs locais estruturados; nunca na apresentação.
type Error struct {
	Code  Code
	Op    string // operação estável, ex.: "collection.sync"
	causa error
}

// New cria o erro tipado da aplicação.
func New(code Code, op string, causa error) *Error {
	return &Error{Code: code, Op: op, causa: causa}
}

// Error é o formato de log local; inclui a causa para diagnóstico.
func (e *Error) Error() string {
	if e.causa == nil {
		return fmt.Sprintf("%s: %s", e.Op, e.Code)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Code, e.causa)
}

// Unwrap expõe a causa para errors.Is/As.
func (e *Error) Unwrap() error { return e.causa }

// Apresentavel é a forma segura para UI e assistente: código e operação,
// sem a causa (que pode conter caminhos absolutos ou detalhes de driver).
func (e *Error) Apresentavel() string {
	return fmt.Sprintf("%s (%s)", e.Op, e.Code)
}

// CodeOf extrai o código estável de qualquer erro; erro não tipado é
// interno por definição.
func CodeOf(err error) Code {
	var tipado *Error
	if errors.As(err, &tipado) {
		return tipado.Code
	}
	return CodeInternal
}
