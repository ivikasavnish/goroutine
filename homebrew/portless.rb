class Portless < Formula
  desc "Named localhost URLs for development servers with ngrok integration"
  homepage "https://github.com/ivikasavnish/goroutine"
  version "1.0.0"
  license "MIT"

  # NOTE: This is a template formula with placeholder SHA256 values.
  # Before using, update with actual checksums using: ./homebrew/update_formula.sh <version>
  # DO NOT use this formula with placeholder values - it poses a security risk!

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "portless-darwin-amd64" => "portless" if OS.mac? && Hardware::CPU.intel?
    bin.install "portless-darwin-arm64" => "portless" if OS.mac? && Hardware::CPU.arm?
    bin.install "portless-linux-amd64" => "portless" if OS.linux? && Hardware::CPU.intel?
    bin.install "portless-linux-arm64" => "portless" if OS.linux? && Hardware::CPU.arm?
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/portless version")
  end
end
