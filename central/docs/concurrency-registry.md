# Registro de concorrência do módulo central

Tarefa 6.4 do OpenSpec `add-central-go-platform`: todo ponto que inicia
goroutine no módulo precisa de uma linha aqui. O teste de arquitetura
(`internal/arch`) impede goroutines fora de `internal/platform/concurrency`,
então novos pontos exigem atualizar este registro e o teste no mesmo commit.

| Ponto (arquivo) | Dono | Contexto pai | Origem do limite | Orçamento downstream | Caminho do resultado/erro | Shutdown |
| --- | --- | --- | --- | --- | --- | --- |
| `concurrency.Source` (pipeline.go) | estágio source; fecha o canal que criou | `errgroup.WithContext` do dono da execução | capacidade do canal: `Budgets.ProviderBuffer` | bloqueia no canal de saída (`Send` seleciona `ctx.Done()`) | erro devolvido ao errgroup; primeiro erro cancela o contexto | retorna em `ctx.Done()`; canal fechado pelo próprio dono |
| Workers de `concurrency.FlatPool`/`Pool` (pool.go) | pool que os criou | `errgroup.WithContext` do dono da execução | parâmetro `workers`, derivado de `config.Budgets` (Fetch/Normalize) | capacidade do canal de saída (`Observation`/`NormalizedBuffer`) | erro de `fn` vai ao errgroup e cancela os demais estágios | saem em `ctx.Done()` ou quando o canal de entrada fecha |
| Fechador de canal de `FlatPool` (pool.go) | pool; única goroutine que fecha o canal de saída | `errgroup.WithContext` do dono da execução | 1 por pool (fixo por construção) | nenhum | sempre `nil`; apenas espera o `WaitGroup` interno | termina quando todos os workers do pool terminam |
| Redutor em `publication.gather` (run.go) | dono da execução (`Run`) | mesmo errgroup dos estágios | 1 por execução — redução é single-owner por spec | consome tudo; paralelismo termina aqui | resultado em variável do dono; erro ao errgroup | `Reduce` seleciona `ctx.Done()` e o canal fechado pelos pools |

Regras herdadas do design (decisão 6):

- Nenhum handler HTTP poderá destacar trabalho após a resposta.
- Nenhuma saída de modelo ou dado de runtime pode elevar `workers`/`buffer`;
  os valores vêm exclusivamente de `config.Budgets`, validados em `[1, max]`.
- O dono da execução é o único que chama `g.Wait()` e responde pelo término
  de todas as goroutines que este registro lista.
