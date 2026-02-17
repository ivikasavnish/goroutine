class Portless < Formula
  desc "Named localhost URLs for development servers with ngrok integration"
  homepage "https://github.com/ivikasavnish/goroutine"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_ARM64"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_ARM64"
    else
      url "https://github.com/ivikasavnish/goroutine/releases/download/v1.0.0/portless-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AMD64"
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
