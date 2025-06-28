<div align="center">
  <h1>ScienceOL Studio</h1>
  <p><em>Studio for Researchers with AI-native Calculation and Laboratory Dispatcher</em></p>
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

**PROTIUM** is a visualization workflow application designed for scientific researchers.

In the AIGC (Artificial Intelligence Generated Content) field, many popular application development kits have adopted workflows as their user interface. For example, in the natural language processing field, [Dify (44.5k Stars)](https://github.com/langgenius/dify) and in the computer vision field, [Comfy UI (49.9k Stars)](https://github.com/comfyanonymous/ComfyUI) have both become popular tools in their respective fields in the form of workflows. The workflow format has been widely practiced and recognized as an AIGC tool. In contrast, the complexity in the AI for Science field is higher, which also provides greater freedom and flexibility for the workflow format, making it potentially a standard tool for the daily work of scientific researchers.

## Features

PROTIUM currently supports the following features:

- **Intuitive Interface:** Easily build and manage reusable scientific computation workflows through a web interface, accessible anytime, anywhere.

  ![1-Basic](./protium/public/introduction/poster/1-Basic.gif)

- **Cloud-native Architecture:** Utilize Docker technology for scientific or AI computations without worrying about complex environment configurations. Supports asynchronous parallel tasks to enhance research efficiency.

- **Enhanced Debugging Experience:** Detailed run logs and error prompts help users quickly locate and resolve issues, improving debugging efficiency.

- **Comprehensive Documentation:** PROTIUM provides a supporting documentation website deeply integrated with the workflow, helping users quickly get started and understand product features.

- **SDK Support:** Interact with PROTIUM through the terminal or Python, view workflow lists, or quickly create workflows via scripts, aiding developers in automating task construction.

  ![3-Cli](./protium/public/introduction/poster/3-Cli.gif)

- **Custom Nodes:** Quickly and efficiently customize workflow nodes through JSON and Python scripts to meet personalized research needs.

- **Community Sharing and Template Library:** Share or publish constructed workflows or node templates to the community with one click. Browse all complete node templates in the community or find workflows that meet your needs.

  ![2-Flociety](./protium/public/introduction/poster/2-Flociety.gif)

- **LLM Integration:** Integrate Large Language Model (LLM) technology, allowing users to quickly build workflows that meet their needs through natural language.

## Quick Start

Protium offers the following access methods:

1. Web
2. Local Deployment
   - Using Docker
   - From Source (currently not supported)
3. Bohrium APP (Workflow)

### 1. Web

We recommend accessing the [Protium Website](https://protium.space) or [Protium Workflow](https://workflows.protium.space) to get started immediately. For any help regarding the website usage, please refer to the [Protium Docs](https://docs.protium.space/workflow).

### 2. Local Deployment

The local deployment is an offline version where all data will be stored locally with no network communication.

#### 1. Using Docker [Recommended]

Please visit the [Docker official website](https://docs.docker.com/get-docker/) to download and install Docker.

Choose the appropriate deployment script based on your geographic location:

- For Users in Mainland China

Mainland China users should use the following script, which utilizes Alibaba Cloud (Aliyun) mirror services to optimize download speeds and overall experience:

```bash
mkdir -p PROTIUM
cd PROTIUM
curl -o docker-compose.yml https://protium.space/downloads/docker-compose-cn.yml
curl -o .env https://protium.space/downloads/example.env
docker compose up -d
```

- For International Users

International users should use the following script, which utilizes Docker's official mirror services for optimal performance:

```bash
mkdir -p PROTIUM
cd PROTIUM
curl -o docker-compose.yml https://protium.space/downloads/docker-compose-en.yml
curl -o .env https://protium.space/downloads/example.env
docker compose up -d
```

These scripts will create a new PROTIUM folder and start the service.

Before starting the service, you need to configure environment variables. A sample `.env` file is provided in the PROTIUM folder for easy setup, which you can modify as needed.

#### 2. From Source

Currently not supported.

## License

This project is licensed under the GPL License, Version 3.0. See the [LICENSE](./LICENSE) file for more details.
