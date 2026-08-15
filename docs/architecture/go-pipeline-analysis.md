# Go pipelines e paralelismo limitado

O diagrama `mtga-go-concurrency.puml` descreve a publicacao de um catalogo, nao o
fluxo interativo completo. Seu objetivo e processar provedores independentes com
concorrencia, sem permitir que carga externa crie goroutines, memoria ou conexoes
PostgreSQL sem limite.

## Como ler o fluxo

1. **Claim do job:** o PostgreSQL concede uma lease e uma chave de idempotencia.
   Assim, retry e troca de processo nao publicam a mesma versao duas vezes.
2. **Run owner:** um unico componente cria `context`, deadline e `errgroup`. Ele e
   responsavel por limitar, cancelar e aguardar todo trabalho iniciado.
3. **Fila de provedores:** os provedores aprovados entram em uma fila de capacidade
   fixa. O fetch pool respeita limites do provedor e de rede.
4. **Observacoes:** produtores enviam para um canal limitado. Quando o consumidor
   fica lento, o canal aplica backpressure em vez de acumular memoria.
5. **Normalizacao:** um pool separado executa trabalho de CPU. Separar pools evita
   que latencia de rede e uso de CPU disputem o mesmo limite.
6. **Reducao estavel:** uma unica goroutine ordena, elimina duplicatas e produz a
   representacao canonica. Essa fronteira preserva bytes e hashes reproduziveis.
7. **Persistencia:** lotes respeitam o orcamento de `pgxpool`; o snapshot e gravado
   como objeto imutavel, verificado e referenciado por um manifesto assinado.
8. **Ativacao:** uma transacao curta muda o manifesto corrente. Leitores nunca
   observam uma publicacao parcialmente concluida.

## O que bounded parallelism significa

Paralelismo limitado e uma politica de capacidade, nao apenas `N` workers. Cada
recurso possui seu proprio orcamento: requisicoes por provedor, CPU, canais,
conexoes de banco e uploads. O menor limite a jusante controla a vazao do fluxo.

Os limites devem ser configuraveis, possuir valor conservador e ser medidos. Um
numero maior pode reduzir o tempo de um job, mas tambem aumenta memoria, disputa
por conexoes e risco de rate limit. Goroutines sao baratas, mas nao gratuitas e
nao devem ser usadas como fila duravel.

## Contrato de implementacao

- Todo `go func` tem owner, politica de termino e `Wait` correspondente.
- O primeiro erro cancela o contexto compartilhado.
- Todo envio bloqueante seleciona entre o canal e `ctx.Done()`.
- Apenas o produtor owner fecha um canal; consumidores nunca o fecham.
- O heartbeat renova a lease enquanto o job esta vivo.
- Erros sao classificados em retryable, terminal ou cancelled.
- Retry usa backoff com jitter e preserva a mesma chave de idempotencia.
- Filas entre processos permanecem no PostgreSQL; canais existem apenas no run.
- A reducao final nao e paralela quando a ordem altera o artefato publicado.

## Como escolher os limites

Comece pelo gargalo externo: limite do provedor, `pgxpool.MaxConns` e banda do
object storage. Reserve conexoes para trafego interativo e use apenas a fracao
restante nos workers. Ajuste com p95 de latencia, utilizacao do pool, tamanho das
filas, memoria do processo, retries e tempo total do job.

O teste de carga deve demonstrar tres propriedades: memoria estabiliza sob produtor
mais rapido, conexoes nunca excedem o orcamento e cancelamento encerra todas as
goroutines dentro do prazo de shutdown.
