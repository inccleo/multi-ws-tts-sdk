#!/bin/bash

# Multi-Context WebSocket TTS Go SDK 发布脚本

set -e

echo "🚀 Go SDK 发布助手"
echo ""

# 检查是否已经是 Git 仓库
if [ -d ".git" ]; then
    echo "⚠️  检测到已存在 Git 仓库"
    read -p "是否继续？这将添加新的提交 (y/N): " confirm
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        echo "已取消"
        exit 0
    fi
    EXISTING_REPO=true
else
    EXISTING_REPO=false
fi

# 获取用户输入
echo ""
echo "📝 请输入仓库信息："
echo ""

read -p "Git 托管平台 (github/gitlab) [github]: " platform
platform=${platform:-github}

read -p "用户名/组织名: " username
if [ -z "$username" ]; then
    echo "❌ 用户名不能为空"
    exit 1
fi

read -p "仓库名 [multi-ws-tts-sdk]: " repo_name
repo_name=${repo_name:-multi-ws-tts-sdk}

read -p "版本号 [v1.0.0]: " version
version=${version:-v1.0.0}

# 构建完整路径
if [ "$platform" = "gitlab" ]; then
    MODULE_PATH="gitlab.com/$username/$repo_name"
    REPO_URL="https://gitlab.com/$username/$repo_name.git"
else
    MODULE_PATH="github.com/$username/$repo_name"
    REPO_URL="https://github.com/$username/$repo_name.git"
fi

echo ""
echo "📦 配置信息："
echo "   模块路径: $MODULE_PATH"
echo "   仓库地址: $REPO_URL"
echo "   版本号: $version"
echo ""

read -p "确认以上信息正确？ (y/N): " confirm
if [[ ! $confirm =~ ^[Yy]$ ]]; then
    echo "已取消"
    exit 0
fi

# 更新 go.mod
echo ""
echo "📝 更新 go.mod..."
sed -i.bak "s|module github.com/yourcompany/multi-ws-tts-sdk|module $MODULE_PATH|g" go.mod
rm go.mod.bak

# 更新示例代码
echo "📝 更新示例代码导入路径..."
find examples -name "*.go" -type f -exec sed -i.bak "s|github.com/yourcompany/multi-ws-tts-sdk|$MODULE_PATH|g" {} \;
find examples -name "*.bak" -type f -delete

# 运行测试
echo ""
echo "🧪 运行测试..."
if go test ./...; then
    echo "✅ 测试通过"
else
    echo "❌ 测试失败，请修复后重试"
    exit 1
fi

# 清理依赖
echo ""
echo "🧹 清理依赖..."
go mod tidy

if [ "$EXISTING_REPO" = false ]; then
    # 初始化 Git 仓库
    echo ""
    echo "📦 初始化 Git 仓库..."
    git init
    git branch -M main
fi

# 添加文件
echo ""
echo "📝 添加文件..."
git add .

# 提交
echo ""
echo "💾 创建提交..."
git commit -m "feat: 初始化 Multi-Context WebSocket TTS Go SDK

- 实现 WebSocket 客户端
- 支持多上下文并发
- 提供 simple 和 multi_context 示例
- 兼容 camelCase 和 snake_case 字段格式
- 版本: $version"

if [ "$EXISTING_REPO" = false ]; then
    # 添加远程仓库
    echo ""
    echo "🔗 添加远程仓库..."
    git remote add origin $REPO_URL
fi

# 推送
echo ""
echo "📤 推送到远程仓库..."
read -p "现在推送到远程仓库？ (y/N): " push_confirm
if [[ $push_confirm =~ ^[Yy]$ ]]; then
    git push -u origin main
    
    # 创建标签
    echo ""
    echo "🏷️  创建版本标签 $version..."
    git tag $version
    git push origin $version
    
    echo ""
    echo "✅ 发布完成！"
    echo ""
    echo "📚 用户可以通过以下方式安装："
    echo "   go get $MODULE_PATH@$version"
    echo ""
    echo "📖 文档将在几分钟后出现在："
    echo "   https://pkg.go.dev/$MODULE_PATH"
else
    echo ""
    echo "⏸️  已准备就绪，你可以稍后手动推送："
    echo "   git push -u origin main"
    echo "   git tag $version"
    echo "   git push origin $version"
fi

echo ""
echo "🎉 完成！"
