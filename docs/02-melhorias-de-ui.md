# Melhorias de UI e UX.

## Parte 1 - Comando '/multiline'

Ao rodar este comando, o usuário visualiza uma mensagem dizendo
que ele pode digitar/colar longas sequências de entrada e que
poderá encerrar esse input com o texto '/send'. Esse final será
detectado quando a linha começa com '/send' e termina com o ENTER.

## Parte 2 - Desativar cores por um flag ou parâmetro

Criar uma nova flag que irá desativar a exibição de cores.
Ao fazer esse refactor, vamos criar uma "paleta de cores" com variáveis
para que possamos configurar melhor depois. Assim, podemos ter uma "paleta zero"
que não aplica cor alguma e apenas imprime o texto.

## Parte 3 - Melhorar o multiline input

Atualmente não há uma distinção entre o usuário digitar enter, linha por linha,
e o usuário copiar e colar. Ao colar, o que ocorre é que o terminal exibe o
prompt de multilne "  >" várias vezes, em sequência. Temos que refatorar o input
handler e tentar detectar esse boost de input e então implementar uma UI mais
limpa no terminal.

## Parte 4 Modo "stdin"

Neste modo, o Zé irá detectar se está em um TTY interativo.
Isso permite criar pequenos "scripts" para implementar alguns
comandos e prompts e o zé irá se comportar da seguinte forma:

1. O banner de boas vindas não é exibido
2. A medida em que for processando comandos em 'stdin', ele irá ecoá-los
   na stdout prefixado com o texto "prompt > "
3. As cores aqui serão desativadas e apenas o plain text será exibido

Desta forma, quando rodarmos o zé com um input como este:

    /multiline
    Escreva um código completo em Go que realizará cálculos matemáticos como o BC:
    - Ele deve suportar inputs como estes: 1+2 ; 5.1 + 10 ; 2*1024
    - Deve suportar apenas as 4 operações, com apenas dois operandos
    /send

Ele irá realizar o prompt multiline, criar o código e sair de forma "limpa" no EOF.
