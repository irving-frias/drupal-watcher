#!/usr/bin/env php
<?php
/**
 * Drupal Watcher — binary downloader and vendor/bin symlink manager
 *
 * Called via:
 *   - Composer post-install-cmd / post-update-cmd
 *   - bin/drupal-watcher shell launcher (first-run download)
 *
 * Scenarios handled:
 *   - composer install               → download binary, create symlink
 *   - composer update                → download binary (if version changed), recreate symlink
 *   - composer remove drupal-watcher → clean up vendor/bin/drupal-watcher
 *   - manual deletion of vendor/     → re-download, recreate symlink
 */

$installDir = __DIR__;
$packageJson = $installDir . '/../composer.json';
$binaryPath = $installDir . '/drupal-watcher-go';
$launcherPath = $installDir . '/drupal-watcher';

// Find project root (directory containing vendor/bin) by walking up
$vendorBin = null;
$dir = $installDir;
for ($i = 0; $i < 10; $i++) {
	$candidate = realpath($dir . '/vendor/bin');
	if ($candidate !== false && is_dir($candidate)) {
		$vendorBin = $candidate;
		break;
	}
	$parent = dirname($dir);
	if ($parent === $dir) break;
	$dir = $parent;
}

// ─── Step 1: if package is not installed, clean up vendor/bin ──────────────
if (!file_exists($packageJson)) {
	if ($vendorBin) {
		$linkPath = $vendorBin . '/drupal-watcher';
		if (is_link($linkPath) || file_exists($linkPath)) {
			@unlink($linkPath);
			echo "🧹 Cleaned up vendor/bin/drupal-watcher (package removed)\n";
		}
	}
	if (file_exists($binaryPath)) {
		@unlink($binaryPath);
	}
	exit(0);
}

// ─── Step 2: package exists — ensure symlink in vendor/bin ─────────────────
$packageRoot = realpath($installDir . '/..');
if ($vendorBin) {
	$linkPath = $vendorBin . '/drupal-watcher';
	$target = $launcherPath;

	// Fix: if symlink points somewhere else or is dead, recreate it
	if (is_link($linkPath) || file_exists($linkPath)) {
		if (!is_link($linkPath) || readlink($linkPath) !== $target) {
			@unlink($linkPath);
		}
	}
	if (!file_exists($linkPath)) {
		@symlink($target, $linkPath);
	}
}

// Read version from our composer.json
$composerMeta = json_decode(file_get_contents($packageJson), true);
$expectedVersion = $composerMeta['extra']['drupal-watcher-version'] ?? '1.0.0';

// ─── Step 3: check if binary needs download ────────────────────────────────
$needsDownload = true;

if (file_exists($binaryPath) && is_executable($binaryPath)) {
	// Check version marker — skip download if same version already installed
	$versionFile = $installDir . '/.binary-version';
	if (file_exists($versionFile) && trim(file_get_contents($versionFile)) === $expectedVersion) {
		$needsDownload = false;
	}
}

if (!$needsDownload) {
	exit(0);
}

// ─── Step 4: download binary from GitHub Releases ──────────────────────────
$repo = 'irving-frias/drupal-watcher';

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
	$version = $expectedVersion;
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
		rename($installDir . '/drupal-watcher-windows-amd64.exe', $binaryPath);
		$zip->close();
	}
	unlink($zipPath);
} else {
	$decompressed = gzdecode($data);
	if ($decompressed === false) {
		echo "✖ Failed to decompress {$archiveName}\n";
		exit(1);
	}
	file_put_contents($binaryPath, $decompressed);
}

chmod($binaryPath, 0755);

// Write version marker so we skip re-download on subsequent runs
file_put_contents($installDir . '/.binary-version', $version);

echo "✔ Installed drupal-watcher {$version} ({$goos}/{$goarch})\n";
