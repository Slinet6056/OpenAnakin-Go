# OpenAnakin-Go

OpenAnakin-Go 是一个兼容 OpenAI API 的 Anakin AI 适配器。它允许使用 OpenAI API 客户端与 Anakin AI 服务进行交互。

## 功能

- 允许使用 OpenAI API 格式调用 Anakin AI 接口
- 支持流式和非流式响应
- 可配置的模型到 Anakin 应用 ID 的映射

## 配置

在 `config.yaml` 文件中配置模型和对应的 Anakin 应用 ID：

```yaml
models:
  gpt-4o-mini: 31800
  gpt-4o: 32442
```

## 使用方法

1. 克隆仓库

2. 配置 config.yaml

3. 运行服务器：

    ```sh
    go run cmd/server/main.go
    ```

4. 在与 OpenAI API 兼容的客户端中，设置 base URL 为 http://localhost:8080/v1，并填入 API 密钥即可使用

## 构建可执行文件

要构建可执行文件，请按照以下步骤操作：

1. 确保您的系统上安装了 Go

2. 在项目根目录下打开终端

3. 运行以下命令：

    ```sh
    go build -o openanakin-go cmd/server/main.go
    ```

4. 构建完成后，可以使用以下命令运行：

    ```sh
    ./openanakin-go
    ```

## Docker 运行

### 从 GitHub Container Registry 运行

1. 拉取镜像：

    ```sh
    docker pull ghcr.io/slinet6056/openanakin-go:master
    ```

2. 准备配置文件：
   创建 `config.yaml` 文件并配置模型映射

3. 运行容器：

    ```sh
    docker run -d \
      -p 8080:8080 \
      -v $(pwd)/config.yaml:/app/config.yaml \
      ghcr.io/slinet6056/openanakin-go:master
    ```

### 本地构建运行

1. 构建镜像：

    ```sh
    docker build -t openanakin-go .
    ```

2. 运行容器：

    ```sh
    docker run -d \
      -p 8080:8080 \
      -v $(pwd)/config.yaml:/app/config.yaml \
      openanakin-go
    ```

服务将在 http://localhost:8080 启动。请确保 config.yaml 文件已正确配置。