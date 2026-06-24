#!/usr/bin/env php
<?php
/**
 * Drupal Watcher — binary downloader
 *
 * Detects OS/arch and downloads the matching pre-built binary from GitHub Releases.
 * Called via Composer post-install-cmd.
 */

$version = getenv('DRUPAL_WATCHER_VERSION') ?: 'v1.0.0-beta3';
$repo = 'irving-frias/drupal-watcher';

// Map PHP_OS to Goos
$osMap = [
    'Linux'  => 'linux',
    'Darwin' => 'darwin',
    'WINNT'  => 'windows',
    'CYGWIN' => 'windows',
    'FreeBSD' => 'freebsd',
];
$goos = $osMap[PHP_OS] ?? strtolower(PHP_OS);

// Map architecture
$archRaw = php_uname('m');
$archMap = [
    'x86_64'  => 'amd64',
    'amd64'   => 'amd64',
    'aarch64' => 'arm64',
    'arm64'   => 'arm64',
    'x86'     => '386',
    'i386'    => '386',
    'i686'    => '386',
];
$goarch = $archMap[$archRaw] ?? $archRaw;

$suffix = $goos === 'windows' ? '.exe' : '';
$binaryName = "drupal-watcher-{$goos}-{$goarch}{$suffix}";

$installDir = __DIR__;
$targetPath = $installDir . '/drupal-watcher-go';

// Already installed?
if (file_exists($targetPath) && is_executable($targetPath)) {
    echo "✔ Drupal Watcher binary already exists.\n";
    exit(0);
}

$url = "https://github.com/{$repo}/releases/download/{$version}/{$binaryName}";
echo "⬇  Downloading {$binaryName}...\n";

$context = stream_context_create([
    'http' => [
        'method' => 'GET',
        'header' => "Accept: application/octet-stream\r\n",
        'timeout' => 30,
        'follow_location' => 1,
    ],
    'ssl' => [
        'verify_peer' => true,
        'verify_peer_name' => true,
    ],
]);

$binary = @file_get_contents($url, false, $context);
if ($binary === false) {
    $error = error_get_last();
    echo "⚠  Download failed: " . ($error['message'] ?? 'unknown error') . "\n";
    echo "   URL: {$url}\n";
    echo "   Install Go or download manually from: https://github.com/{$repo}/releases\n";
    exit(1);
}

$written = file_put_contents($targetPath, $binary);
if ($written === false) {
    echo "✖ Failed to write binary to {$targetPath}\n";
    exit(1);
}

chmod($targetPath, 0755);
echo "✔ Installed drupal-watcher ({$goos}/{$goarch}) to {$targetPath}\n";
