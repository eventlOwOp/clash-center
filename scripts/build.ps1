#!/usr/bin/env pwsh

# 定义变量
$ProjectName = "clash-center"
$Version = "1.0.0"
$BuildDir = "build"

# 创建构建目录
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

# 定义目标平台和架构
$Targets = @(
    @{OS = "linux"; Arch = "amd64"; File = "${ProjectName}-linux-amd64"},
    @{OS = "linux"; Arch = "arm64"; File = "${ProjectName}-linux-arm64"},
    @{OS = "linux"; Arch = "arm"; Arm = "7"; File = "${ProjectName}-linux-armv7"}
)

# 显示开始编译消息
Write-Host "开始为 $ProjectName 构建跨平台二进制文件..." -ForegroundColor Cyan

# 遍历目标平台进行编译
foreach ($Target in $Targets) {
    $OS = $Target.OS
    $Arch = $Target.Arch
    $OutputFile = Join-Path -Path $BuildDir -ChildPath $Target.File
    
    # 设置环境变量
    $Env:GOOS = $OS
    $Env:GOARCH = $Arch
    
    # 如果是ARM架构，设置ARM版本
    if ($Arch -eq "arm" -and $Target.Arm) {
        $Env:GOARM = $Target.Arm
        Write-Host "正在构建: $OS/$Arch v$($Target.Arm) -> $OutputFile" -ForegroundColor Yellow
    } else {
        Write-Host "正在构建: $OS/$Arch -> $OutputFile" -ForegroundColor Yellow
    }
    
    # 执行Go构建命令
    go build -trimpath -ldflags "-s -w -X main.Version=$Version" -o "$OutputFile" .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  编译成功: $OutputFile" -ForegroundColor Green
    } else {
        Write-Host "  编译失败: $OS/$Arch" -ForegroundColor Red
    }
}

Write-Host "所有编译任务完成。文件保存在 $BuildDir 目录中。" -ForegroundColor Cyan 