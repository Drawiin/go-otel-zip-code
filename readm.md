```markdown
## Instruções para Executar o Projeto

Este projeto requer o Docker e o Docker Compose para ser executado, além de uma chave de API válida para o serviço de clima.

### Pré-requisitos

- Docker/Docker Compose: Certifique-se de que o Docker e o Docker Compose estão instalados em seu sistema.
- Chave de API do Serviço de Clima: Obtenha uma chave de API válida para acessar o serviço de clima.

### Passos para Executar

#### Usando Docker

1. **Clone o Repositório**: Primeiro, clone o repositório do projeto em sua máquina local:
   ```sh
   git clone <URL-do-repositório>
   cd <nome-do-repositório>
   ```

2. **Configure a Chave de API**: Crie um arquivo `.env` na raiz do projeto e adicione sua chave de API:
   ```sh
   WEATHER_API_KEY=<sua-chave-de-api>
   ```

3. **Construa e Inicie os Containers Docker**: Use o Docker Compose para construir e iniciar os containers:
   ```sh
   docker-compose up --build
   ```

4. **Acesse o Serviço**: Após iniciar os containers, o serviço estará disponível em `http://localhost:8080`.

### Verificando o Funcionamento

- **Teste com Arquivo Integrado**: Utilize o arquivo `integrated.http` para testar o serviço. Abra o arquivo e clique no botão `Send Request` para enviar uma requisição de exemplo.
- **Ferramentas Alternativas**: Você também pode usar ferramentas como `curl`, `Postman` ou `Insomnia` para testar o serviço manualmente.
- **Verificação de Traces**: Após realizar algumas requisições, você pode verificar os traces gerados acessando o Zipkin em `http://localhost:9411/`.
```