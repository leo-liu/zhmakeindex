# Homebrew Formula 模板
#
# 使用方式：
# 1. 创建你自己的 tap 仓库，如 github.com/leo-liu/homebrew-zhmakeindex
# 2. 将此文件放入 tap 仓库的 Formula/ 目录
# 3. 每次 release 后，更新 version、url 和 sha256
# 4. 用户安装命令：brew install leo-liu/zhmakeindex/zhmakeindex
#
# 自动化更新（可选）：
# 在 release workflow 完成后，用脚本或 CI 自动更新 url 和 sha256：
#   VERSION=x.y.z
#   URL="https://github.com/leo-liu/zhmakeindex/releases/download/v${VERSION}/zhmakeindex_${VERSION}_darwin_arm64.tar.gz"
#   SHA256=$(curl -sL "$URL" | shasum -a 256 | cut -d ' ' -f1)
#   sed -i '' "s|url \".*\"|url \"${URL}\"|; s|sha256 \".*\"|sha256 \"${SHA256}\"|" Formula/zhmakeindex.rb
#
class Zhmakeindex < Formula
  desc "Chinese-aware makeindex replacement for LaTeX"
  homepage "https://github.com/leo-liu/zhmakeindex"
  version "0.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/leo-liu/zhmakeindex/releases/download/v#{version}/zhmakeindex_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_ARM64"
    elsif Hardware::CPU.intel?
      url "https://github.com/leo-liu/zhmakeindex/releases/download/v#{version}/zhmakeindex_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/leo-liu/zhmakeindex/releases/download/v#{version}/zhmakeindex_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"
    elsif Hardware::CPU.intel?
      url "https://github.com/leo-liu/zhmakeindex/releases/download/v#{version}/zhmakeindex_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"
    end
  end

  def install
    bin.install "zhmakeindex"
  end

  test do
    assert_match "zhmakeindex", shell_output("#{bin}/zhmakeindex 2>&1", 0)
  end
end
