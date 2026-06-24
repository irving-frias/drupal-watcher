#!/usr/bin/env php
<?php
/**
 * Drupal Watcher — binary downloader
 *
 * Detects OS/arch and downloads the matching pre-built binary from GitHub Releases.
 * Called via Composer post-install-cmd or bin/drupal-watcher launcher.
 */

$installDir = __DIR__;

$repo = 'irving-frias/drupal-watcher';

// Get latest version from GitHub API (with composer.json fallback)
$version = getenv('DRUPAL_WATCHER_VERSION');
if (!$version) {
  $apiUrl = "https://api.github.com/repos/{$repo}/releases/latest";
  $apiCtx = stream_context_create(['http' => ['header' => "User-Agent: Composer\r\n", 'timeout' => 5]]);
  $apiData = @file_get_contents($apiUrl, false, $apiCtx);
  if ($apiData) {
    $release = json_decode($apiData, true);
    if (isset($release['tag_name'])) {
      $version = $release['tag_name'];
    }
  }
}
if (!$version) {
  $composerJson = json_decode(file_get_contents($installDir . '/../composer.json'), true);
  $version = $composerJson['extra']['drupal-watcher-version'] ?? 'v1.0.0';
}

$targetPath = $installDir . '/drupal-watcher-go';

if (file_exists($targetPath) && is_executable($targetPath)) {
  echo "✔ Drupal Watcher binary already exists.\n";
  exit(0);
}

// Map OS
$osMap = [
  'Linux'    => 'linux',
  'Darwin'   => 'darwin',
  'WINNT'    => 'windows',
  'CYGWIN'   => 'windows',
  'FreeBSD'  => 'freebsd',
];
$goos = $osMap[PHP_OS] ?? strtolower(PHP_OS);

// Map arch
$archMap = [
  'x86_64'  => 'amd64',
  'amd64'   => 'amd64',
  'aarch64' => 'arm64',
  'arm64'   => 'arm64',
  'x86'     => '386',
  'i386'    => '386',
  'i686'    => '386',
];
$goarch = $archMap[php_uname('m')] ?? php_uname('m');

$binaryName = "drupal-watcher-{$goos}-{$goarch}";
$isWindows = $goos === 'windows';

if ($isWindows) {
  $archiveName = $binaryName . '.exe.zip';
} else {
  $archiveName = $binaryName . '.gz';
}

$url = "https://github.com/{$repo}/releases/download/{$version}/{$archiveName}";
echo "⬇  Downloading {$archiveName}...\n";

$context = stream_context_create([
  'http' => [
    'method' => 'GET',
    'timeout' => 30,
    'follow_location' => 1,
    'header' => "Accept: application/octet-stream\r\n",
  ],
  'ssl' => [
    'verify_peer' => true,
    'verify_peer_name' => true,
  ],
]);

$data = @file_get_contents($url, false, $context);
if ($data === false) {
  echo "⚠  Download failed: {$url}\n";
  echo "   Install Go or download manually from: https://github.com/{$repo}/releases\n";
  exit(1);
}

if ($isWindows) {
  if (!class_exists('ZipArchive')) {
    echo "✖ Zip extension required. Unzip manually: {$url}\n";
    exit(1);
  }
  $zipPath = $installDir . '/drupal-watcher-tmp.zip';
  file_put_contents($zipPath, $data);
  $zip = new ZipArchive;
  if ($zip->open($zipPath) === true) {
    $zip->extractTo($installDir, ['drupal-watcher-windows-amd64.exe']);
    rename($installDir . '/drupal-watcher-windows-amd64.exe', $targetPath);
    $zip->close();
  }
  unlink($zipPath);
} else {
  $decompressed = gzdecode($data);
  if ($decompressed === false) {
    echo "✖ Failed to decompress {$archiveName}\n";
    exit(1);
  }
  file_put_contents($targetPath, $decompressed);
}

chmod($targetPath, 0755);
echo "✔ Installed drupal-watcher ({$goos}/{$goarch}) to {$targetPath}\n";

// Ensure vendor/bin/drupal-watcher symlink exists (from vendor context)
$vendorBin = realpath($installDir . '/../../../bin');
if ($vendorBin !== false && is_dir($vendorBin)) {
	$launcher = $installDir . '/drupal-watcher';
	$linkPath = $vendorBin . '/drupal-watcher';
	if (!file_exists($linkPath)) {
		@symlink($launcher, $linkPath);
	}
}
