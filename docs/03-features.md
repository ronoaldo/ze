# Novas Features

Plano de funcionalidades desejáveis para o Zé! Esse plano deve ser implementado com TDD.

## Feature 1: Executar comandos no shell com !

Status: Done

A primeira feature desejada é executar comandos com o !, como "!ls",
e ver o stdout/stderr deste comando na sessão do zé.

O agente deve então acumular os comandos executados até o próximo prompt ser enviado para a IA.
Ao enviar, os comandos executados devem ir para o modelo LLM, como mostrado no exemplo abaixo.

Supondo esta sessão de interação com o zé:

```
ze > !ls
file1.txt

ze > Analise o conteúdo do arquivo listado
```

A ia recebe a combinação do prompt do usuário e o comando que ele rodou:

```
User executed command:
$ ls
file1.txt

Analise o conteúdo do arquivo listado.
```

## Feature 2: Tool WebFetch para baixar da web

Status: Done

Essa tool é o mínimo até termos WebSearch via MCPs.

Essa será a tool que pode ser usada pelo modelo quando o usuário passa
links contendo páginas de documentação, etc.

Ela deve, por segurança, só baixar incluir no contexto formato texto:
HTML, JSON, Markdown, TXT, CSV serão aceitos.

## Feature 3: Renderizar Markdown da resposta no terminal

Status: Done

Renderizador básico para suportar:
* Tabelas
* Listas
* Cabeçalhos e Negrito
* Itálico
* Sublinhado

## Feature 4: Tool de "ver a doc completa" go_doc('all')

Status: Done

Melhorar a tool para rodar algo equivalente a:

    go list ./... | while read pkg ; do go doc -all $pkg ; done

Desta forma conseguimos ter um "full api" da implementação atual.

Status: Done

## Feature 5: Nova ferramenta: git_commit

Status: Done

Esta ferramenta irá receber a mensagem do commit. Ela deve ter na
descrição de forma bem explícita que só pode ser feita DEPOIS da
aprovação explícita do usuário para fazer o commit. Ela não deve
ser chamada apenas para gerar a mensagem. Sempre confirmar com o usuário antes.

## Feature 6: Histórico de sessão persistente

Status: Done

Criar um UUID ao iniciar uma sessão, que ficará salva em $ZE_HOME/sessions,
que por padrão é $HOME/.config/ze.

Persistir o histórico de conversa em um arquivo JSON. Atualizar o arquivo a cada
iteração (a cada input de usuário, cada turno novo da AI).

A ideia é salvar em um formato que pode ser lido diretamente para o histórico do Zé:
ao iniciar com --session=UUID, recarregar o histórico e no próximo turno do usuário
todo o contexto carregado será então enviado.

## Feature 7: Logs de execução em arquivo

Status: Done

Gravar em um arquivo de logs, por padrão $ZE_HOME/logs, ZE_HOME=$HOME/.config/ze
por padrão.

Os logs devem detalhar a execução e ter o req/resp de cada chamada de API logados
também para que possamos debugar a execução.

Na primeira versão, é suficiente ter a implementação persistindo somente o req/resp
e os timings.
