package devmcp

import (
	"context"
	"encoding/json"
)

type callParams struct {
	Name      string `json:"name"`
	Arguments struct {
		Scenario string          `json:"scenario"`
		RunID    string          `json:"run_id"`
		Fixture  string          `json:"fixture"`
		Event    json.RawMessage `json:"event"`
	} `json:"arguments"`
}

func textResult(text string) any {
	return map[string]any{"content": []map[string]string{
		{"type": "text", "text": text}}}
}

func toolError(message string) *rpcError {
	return &rpcError{Code: -32000, Message: message}
}

func (s *Server) call(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var params callParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}
	switch params.Name {
	case "graphs_list":
		if s.Deps.ListGraphs == nil {
			return nil, toolError("graph source not attached")
		}
		names, err := s.Deps.ListGraphs(ctx)
		if err != nil {
			return nil, toolError(err.Error())
		}
		return textResult(s.Lim.withDefaults().page(names)), nil
	case "scenario_execute":
		if !s.Caps.ScenarioExecute {
			return nil, &rpcError{Code: -32003, Message: "read-only by default: scenario-execute capability not granted"}
		}
		if s.Deps.ExecuteScenario == nil {
			return nil, toolError("scenario execution not attached")
		}
		if params.Arguments.Scenario == "" {
			return nil, &rpcError{Code: -32602, Message: "scenario is required"}
		}
		if !s.Caps.allowsScenario(params.Arguments.Scenario) {
			return nil, &rpcError{Code: -32003, Message: "scenario outside every isolated namespace"}
		}
		report, err := s.Deps.ExecuteScenario(ctx, params.Arguments.Scenario)
		if err != nil {
			return nil, toolError(err.Error())
		}
		return textResult(report), nil
	case "run_timeline":
		if s.Deps.RunTimeline == nil {
			return nil, toolError("timeline source not attached")
		}
		lines, err := s.Deps.RunTimeline(ctx, params.Arguments.RunID)
		if err != nil {
			return nil, toolError(err.Error())
		}
		return textResult(s.Lim.withDefaults().page(lines)), nil
	case "events_validate":
		if s.Deps.ValidateEvent == nil {
			return nil, toolError("event contract not attached")
		}
		if err := s.Deps.ValidateEvent(ctx, params.Arguments.Event); err != nil {
			return textResult("invalid: " + err.Error()), nil
		}
		return textResult("valid"), nil
	case "fixtures_diagnose":
		if s.Deps.DiagnoseFixture == nil {
			return nil, toolError("fixture source not attached")
		}
		report, err := s.Deps.DiagnoseFixture(ctx, params.Arguments.Fixture)
		if err != nil {
			return nil, toolError(err.Error())
		}
		return textResult(report), nil
	default:
		return nil, &rpcError{Code: -32602, Message: "unknown tool"}
	}
}
