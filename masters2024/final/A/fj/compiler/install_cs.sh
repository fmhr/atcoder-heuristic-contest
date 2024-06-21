#!/bin/bash

# Microsoft パッケージのダウンロードとインストール
wget -q https://packages.microsoft.com/config/ubuntu/22.10/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
dpkg -i packages-microsoft-prod.deb
rm packages-microsoft-prod.deb

# dotnet パッケージの優先度設定
sh -c "cat > /etc/apt/preferences.d/dotnet <<'EOF'
Package: dotnet*
Pin: origin packages.microsoft.com
Pin-Priority: 1001
EOF"
sh -c "cat > /etc/apt/preferences.d/aspnet <<'EOF'
Package: aspnet*
Pin: origin packages.microsoft.com
Pin-Priority: 1001
EOF"

# SDK のインストール
# 細かいversionを指定するとエラーがでるので、7.0にしておく
apt-get update && apt-get install -y dotnet-sdk-7.0 clang zlib1g-dev

# カレントディレクトリーにプロジェクトファイルを設置
cat > Main.csproj << EOS
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net7.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>annotations</Nullable>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
    <DefineConstants>ONLINE_JUDGE;ATCODER</DefineConstants>
    <PublishAot>true</PublishAot>
    <IlcOptimizationPreference>Speed</IlcOptimizationPreference>
    <SatelliteResourceLanguages>en-US</SatelliteResourceLanguages>
    <InvariantGlobalization>true</InvariantGlobalization>
    <WarningsAsErrors>IL2104;IL3053</WarningsAsErrors>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="ac-library-csharp" Version="3.0.0-atcoder8" />
    <PackageReference Include="MathNet.Numerics" Version="5.0.0" />
  </ItemGroup>
</Project>
EOS

# プロジェクトをリストア。あらかじめコンパイルを通してWJの短縮を試みる
echo 'Console.WriteLine("Hello, world!");' > Main.cs
export DOTNET_EnableWriteXorExecute=0
dotnet publish -c Release -o tmp -v q --nologo 1>&2
rm Main.cs

# ローカル環境M1Macでテストするために、以下を外している
# <RuntimeIdentifier>ubuntu-x64</RuntimeIdentifier>