package workflow

import "fmt"

// validar aplica todas as regras estruturais da tarefa 2.2: identidade,
// nós referenciados, alcançabilidade, caminho terminal e limites nos tetos.
func validar(d Definition, caps Limits) error {
	if d.Identity.Kind == "" || d.Identity.Version < 1 {
		return fmt.Errorf("%w: identidade %s", ErrInvalidDefinition, d.Identity)
	}
	if err := d.Recovery.valida(); err != nil {
		return err
	}
	if len(d.Nodes) == 0 {
		return fmt.Errorf("%w: %s sem nós", ErrInvalidDefinition, d.Identity)
	}
	if _, ok := d.Nodes[d.Initial]; !ok {
		return fmt.Errorf("%w: nó inicial %q inexistente", ErrInvalidDefinition, d.Initial)
	}
	if err := validarTransicoes(d); err != nil {
		return err
	}
	if err := validarAlcancabilidade(d); err != nil {
		return err
	}
	return validarLimites(d.Identity, d.Limits, caps)
}

// validarTransicoes garante que toda transição referencia nós existentes e
// que todo nó tem pelo menos uma saída (para nó ou terminal).
func validarTransicoes(d Definition) error {
	saidas := make(map[NodeID]int, len(d.Nodes))
	for chave, alvo := range d.Transitions {
		if _, ok := d.Nodes[chave.From]; !ok {
			return fmt.Errorf("%w: transição de nó inexistente %q", ErrInvalidDefinition, chave.From)
		}
		if alvo.EhTerminal() == (alvo.Next != "") {
			return fmt.Errorf("%w: alvo de %q/%q deve ter exatamente Next ou Terminal",
				ErrInvalidDefinition, chave.From, chave.Outcome)
		}
		if !alvo.EhTerminal() {
			if _, ok := d.Nodes[alvo.Next]; !ok {
				return fmt.Errorf("%w: transição para nó inexistente %q", ErrInvalidDefinition, alvo.Next)
			}
		}
		saidas[chave.From]++
	}
	for id := range d.Nodes {
		if saidas[id] == 0 {
			return fmt.Errorf("%w: nó %q sem transição de saída", ErrInvalidDefinition, id)
		}
	}
	return nil
}

// validarAlcancabilidade exige que todo nó seja alcançável a partir do
// inicial e que exista pelo menos um outcome terminal alcançável.
func validarAlcancabilidade(d Definition) error {
	visitados := map[NodeID]bool{d.Initial: true}
	fila := []NodeID{d.Initial}
	terminalAlcancavel := false
	for len(fila) > 0 {
		atual := fila[0]
		fila = fila[1:]
		for chave, alvo := range d.Transitions {
			if chave.From != atual {
				continue
			}
			if alvo.EhTerminal() {
				terminalAlcancavel = true
			} else if !visitados[alvo.Next] {
				visitados[alvo.Next] = true
				fila = append(fila, alvo.Next)
			}
		}
	}
	for id := range d.Nodes {
		if !visitados[id] {
			return fmt.Errorf("%w: nó %q inalcançável", ErrInvalidDefinition, id)
		}
	}
	if !terminalAlcancavel {
		return fmt.Errorf("%w: %s sem outcome terminal alcançável", ErrInvalidDefinition, d.Identity)
	}
	return nil
}

// validarLimites permite reduzir limites, nunca exceder os tetos duros.
func validarLimites(id Identity, l, caps Limits) error {
	if l.MaxSteps > caps.MaxSteps || l.MaxToolCalls > caps.MaxToolCalls ||
		l.MaxRepeatedCalls > caps.MaxRepeatedCalls ||
		l.MaxPayloadBytes > caps.MaxPayloadBytes || l.ActiveDeadline > caps.ActiveDeadline {
		return fmt.Errorf("%w: %s excede tetos duros", ErrInvalidDefinition, id)
	}
	return nil
}
