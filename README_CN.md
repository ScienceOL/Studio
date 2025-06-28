<div align="center">
  <h1>ScienceOL</h1>
  <p><em>为科学研究设计的可视化工作流应用程序</em></p>
  <p>
    <a href="./README.md">English</a> |
    <a href="./README_CN.md">简体中文</a>
  </p>
</div>

## 编辑器配置

我们建议使用 Visual Studio Code 进行开发，并已在仓库中包含了推荐的 VS Code 设置。

### VS Code 设置

项目的 `.vscode/settings.example.json` 文件已包含正确的配置。使用以下命令创建工作区的配置文件。

```bash
cp .vscode/settings.example.json .vscode/settings.json
```

VSCode 应该会自动应用这些设置。

---

**PROTIUM** 是一款专为科学研究人员设计的可视化工作流应用程序。

在 AIGC 领域，许多热门应用开发套件都采用了工作流作为用户界面。例如，自然语言处理领域的 [Dify (44.5k Star)](https://github.com/langgenius/dify) 和计算机视觉领域的 [Comfy UI(49.9k Star)](https://github.com/comfyanonymous/ComfyUI) 都以工作流形式成为了各自领域的热门工具。工作流作为 AIGC 工具的形式已经被广泛实践并认可。相比之下，AI for Science 领域的复杂度更高，这也为工作流形式提供了更大的自由度和灵活性，使其有潜力成为科学研究人员日常工作的标准工具。

## 特性

PROTIUM 将支持以下特性：

- **可视化界面**：通过 Web 界面轻松构建和管理可复用的科学计算工作流，随时随地一键上手。

  ![1-Basic](./protium/public/introduction/poster/1-Basic.gif)

- **云原生架构**：利用 Docker 技术进行科学或 AI 计算，无需担心复杂的环境配置。支持异步并行任务，提升科研效率。

- **优化调试体验**：详细的运行日志和错误提示，帮助用户迅速定位和解决问题，提升调试效率。

- **完善文档支持**：PROTIUM 提供配套的文档网站，与工作流深度集成，帮助用户快速上手并深入了解产品功能。

- **SDK 支持**：支持通过终端或 Python 与 PROTIUM 互动，查看工作流列表或通过脚本快速创建工作流，助力开发者实现自动化任务构建。

  ![3-Cli](./protium/public/introduction/poster/3-Cli.gif)

- **自定义节点**：通过 JSON 和 Python 脚本，用户可以快速高效地自定义工作流节点，满足个性化的科研需求。

- **社区共享和模板库**：用户可以一键与他人共享或发布构建的工作流或节点模板到社区。在社区中浏览所有完整的节点模板，或寻找符合需求的工作流。

  ![2-Flociety](./protium/public/introduction/poster/2-Flociety.gif)

- **LLM 集成**：集成大语言模型（LLM）技术，用户可以通过自然语言快速构建符合需求的工作流。

## 快速开始

Protium 提供以下三种访问方式：

1. Web
2. 本地部署
   - 使用 Docker 部署
   - 使用源码部署（暂不支持）
3. Bohrium APP(工作流)

### 1. Web

推荐访问 [Protium 网站](https://protium.space) 或 [Protium Workflow](https://workflows.protium.space) 立即开始。有关网站的任何使用帮助，请参阅 [Protium Docs](https://docs.protium.space/workflow)。

### 2. 本地部署

本地部署为离线版本，所有数据将保存在你的本地，不发生网络通信。

#### 1. 使用 Docker 部署【推荐】

请访问 [Docker 官方网站](https://docs.docker.com/get-docker/) 下载并安装 Docker。

根据您的地理位置选择合适的部署脚本：

- 适用于中国大陆用户

中国大陆用户请使用以下脚本，该脚本使用阿里云镜像服务来优化下载速度和整体体验：

```bash
mkdir -p PROTIUM
cd PROTIUM
curl -o docker-compose.yml https://protium.space/downloads/docker-compose-cn.yml
curl -o .env https://protium.space/downloads/example.env
docker compose up -d
```

- 适用于国际用户

国际用户请使用以下脚本，该脚本使用 Docker 官方镜像服务以确保最佳性能：

```bash
mkdir -p PROTIUM
cd PROTIUM
curl -o docker-compose.yml https://protium.space/downloads/docker-compose-en.yml
curl -o .env https://protium.space/downloads/example.env
docker compose up -d
```

以上脚本将新建一个 PROTIUM 文件夹并启动服务。

启动服务前你需要配置环境变量。PROTIUM 文件夹下提供了示例 `.env` 文件便于一键启动服务，如有需要可自行修改。

#### 2. 使用源码部署

暂不支持。

## 许可证

此项目使用 GPLv3 许可证。有关更多详情，请参阅 [LICENSE](./LICENSE) 文件。

<div align="center">
  <h2>开发环境配置</h2>
</div>
1. 本地安装 docker 
2. python 环境使用 3.12.4
3. 启动 ./launch/dev.sh
