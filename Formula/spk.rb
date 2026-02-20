# This formula is auto-updated by GoReleaser on each tagged release.
# Manual edits will be overwritten.

class Spk < Formula
  desc "Workspace CLI for multi-repo development"
  homepage "https://github.com/Spark-Rewards/homebrew-spk"
  license "MIT"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Spark-Rewards/homebrew-spk/releases/download/v0.1.0/spk_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/Spark-Rewards/homebrew-spk/releases/download/v0.1.0/spk_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Spark-Rewards/homebrew-spk/releases/download/v0.1.0/spk_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/Spark-Rewards/homebrew-spk/releases/download/v0.1.0/spk_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    bin.install "spk"
  end

  test do
    system "#{bin}/spk", "version"
  end
end
