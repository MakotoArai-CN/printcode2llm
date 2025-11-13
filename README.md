<div  align="center"> 
<img src="./ptlm.png" width="200" height="200">
<h1>PrintCode2LLM</h1>
</div>

## 更好的面向AI Chat编程

把项目代码整理成 Markdown，方便喂给大模型。项目没有任何破解，仅打印项目代码，大项目不建议使用本项目打印项目（例如总字符超过2 000 000字符）。

> 注意：直接执行`ptlm`默认打印执行目录下的所有代码文件，执行时请注意执行时机。

## 安装

### 下载预编译版本

从 [Releases](https://github.com/MakotoArai-CN/printcode2llm/releases) 下载对应平台的文件。

### 从源码编译

```bash
git clone https://github.com/MakotoArai-CN/printcode2llm.git
cd printcode2llm
make
```

### 安装到系统

```bash
./ptlm install
```

## 基本用法

```bash
ptlm                        # 整理当前目录
ptlm ./myproject            # 整理指定目录
ptlm ./proj1 ./proj2        # 整理多个项目
```

生成的文件：`LLM_CODE_Part1_of_2.md`、`LLM_CODE_Part2_of_2.md` 等。

## 配置

首次运行会生成 `.ptlm.yaml`：

```yaml
custom_ignore:
  patterns:
    - "*.test.go"
    - "tmp/*"
  regex:
    - ".*_backup\\..*"

output:
  max_chars: 50000
  compress: true
  ultra_compress: false
  output_prefix: LLM_CODE
```

### 手动生成配置文件

```bash
ptlm config init
```

## 常用参数

```bash
-c, --chars 30000           # 每段最大字符数
-o, --output MY_CODE        # 输出文件前缀
-u, --ultra-compress        # 超级压缩模式
-f, --config custom.yaml    # 指定配置文件
--exclude "*.test.go,tmp/*" # 排除文件
--regex ".*_test\\.go$"     # 正则排除
--no-tree                   # 不生成目录树
```

## 过滤规则

### 默认忽略

- 依赖：`node_modules`、`vendor`、`venv`
- 版本控制：`.git`、`.svn`
- 构建产物：`dist`、`build`、`target`
- IDE：`.vscode`、`.idea`
- 二进制文件：`.exe`、`.so`、`.dll`

### 自定义排除

命令行：
```bash
ptlm --exclude "test/*,*.bak" --regex ".*_backup\\..*"
```

配置文件：
```yaml
custom_ignore:
  patterns:
    - "test/*"
    - "*.bak"
  regex:
    - ".*_backup\\..*"
```

## 压缩模式

### 标准压缩（默认）

- 删除注释
- 删除空行
- 保留基本格式

### 超级压缩

```bash
ptlm -u
```

- 激进压缩
- 删除几乎所有空白
- 需要格式化才能阅读

## 版本管理

```bash
ptlm version                # 查看版本
ptlm update                 # 检查更新
ptlm uninstall              # 卸载
```

## 编译所有平台

```bash
make                        # 编译所有平台
make windows                # 只编译 Windows
make linux                  # 只编译 Linux
make darwin                 # 只编译 macOS
make compress               # 用 upx 压缩（需要安装 upx）
```

输出在 `build/` 目录。

## 支持的平台

- Windows: 32位、64位、ARM64
- Linux: 32位、64位、ARM、ARM64、龙芯
- macOS: Intel、Apple Silicon
- FreeBSD: 32位、64位、ARM

## 实际例子

### 整理 Go 项目

```bash
ptlm --exclude "vendor/*,*_test.go" -c 40000
```

### 整理前端项目

```bash
ptlm --exclude "node_modules,dist,*.min.js" ./src
```

### 多项目对比

```bash
ptlm ./old-version ./new-version
```

### 使用自定义配置

```bash
ptlm -f project.yaml ./myproject
```

## 目录结构

```
.
├── cmd/ptlm/           # 主程序
├── internal/
│   ├── cli/            # 命令行
│   ├── compress/       # 代码压缩
│   ├── config/         # 配置管理
│   ├── generator/      # 内容生成
│   ├── output/         # 文件输出
│   ├── scanner/        # 文件扫描
│   └── ui/             # 界面输出
├── configs/            # 配置文件
└── Makefile
```

## 各家大模型单次最大输入长度

|                      网站                       |     模型     | 网站类型 |                                最大输入长度                                |
| :---------------------------------------------: | :----------: | :------: | :------------------------------------------------------------------------: |
|        [lmarena.ai](https://lmarena.ai/)        |     全部     |   三方   |                                   119993                                   |
|      [chat.qwen.ai](https://chat.qwen.ai/)      |   Qwen系列   |   官方   |           未知（代码量达到一定量时，发送的代码会以文件形式发送）           |
|    [www.tongyi.com](https://www.tongyi.com/)    |   Qwen系列   |   官方   |           未知（代码量达到一定量时，发送的代码会以文件形式发送）           |
|          [grok.com](https://grok.com)           |   Grok系列   |   官方   |   未知（貌似没有上限，但是单次最好不要超过2 000 000字符，浏览器会卡死）    |
| [chat.deepseek.com](https://chat.deepseek.com/) | Deepseek系列 |   官方   | ~300 000（代码量达到一定量时，直接上限，强制结束当前对话，必须开启新对话） |
|    [www.doubao.com](https://www.doubao.com/)    |    Doubao    |   官方   |                  ~10 000（代码量达到一定量时，直接上限）                   |

> 仅做简单估算，具体请自行测试。
> 
> 以上数据非官方数据，仅供参考。

## 常见问题

**Q: 生成的文件太大了？**

降低字符限制：`ptlm -c 30000`

**Q: 想排除测试文件？**

```bash
ptlm --exclude "*_test.go,test/*"
```

**Q: 压缩后代码看不懂？**

别用超级压缩，或者用格式化工具：
```bash
# JavaScript
npx prettier --write file.js

# Python
black file.py

# Go
gofmt -w file.go
```

**Q: 想自定义提示词？**

编辑 `configs/prompts.yaml` 或 `.ptlm.yaml`。

## License

MIT

## 贡献

欢迎 [PR](https://github.com/MakotoArai-CN/printcode2llm/pulls)。

报 bug 或提建议：[Issues](https://github.com/MakotoArai-CN/printcode2llm/issues)