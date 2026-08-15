package ports

import (
	"context"
	"time"
)

// ConsentState é o estado corrente de consentimento de telemetria: opt-in
// explícito, com propósito e versão registrados (spec consented-telemetry
// do central; o companion nunca envia nada sem isto).
type ConsentState struct {
	OptIn   bool
	Purpose string
	Version int
}

// ConsentStore expõe o consentimento vigente do usuário.
type ConsentStore interface {
	Current(ctx context.Context) (ConsentState, error)
}

// TelemetryEvent é um evento de produto local, candidato a envio. Seq é a
// posição durável na fila local e a base do acknowledgement.
type TelemetryEvent struct {
	Seq   int
	Name  string
	At    time.Time
	Attrs map[string]string
}

// TelemetryQueue é a fila local durável de eventos: pendências em ordem e
// confirmação até um seq — eventos não confirmados permanecem locais.
type TelemetryQueue interface {
	Pending(ctx context.Context, limite int) ([]TelemetryEvent, error)
	Ack(ctx context.Context, ateSeq int) error
}

// TelemetrySink envia um lote idempotente ao central; o adapter real fala
// HTTPS com chave de idempotência, o fake do harness registra e falha
// sob demanda.
type TelemetrySink interface {
	SendBatch(ctx context.Context, lote []TelemetryEvent) error
}
